package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/ini.v1"
)

type MongoConnector struct {
	client *mongo.Client
	db     *mongo.Database
	colls  map[string]*mongo.Collection
}

func (mcon *MongoConnector) getCollection(collectionName string) *mongo.Collection {
	if mcon.colls[collectionName] == nil {
		mcon.colls[collectionName] = mcon.db.Collection(collectionName)
	}
	return mcon.colls[collectionName]
}

func (mcon *MongoConnector) connect() {
	cfg, err := ini.Load(".ini")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in parsing .ini file for mongo credentials\n")
		os.Exit(1)
	}
	var uri string = cfg.Section("mongodb").Key("uri").String()
	var db string = cfg.Section("mongodb").Key("db").String()
	credential := options.Credential{
		AuthSource: db,
		// AuthMechanism: "SCRAM-SHA-256",
		Username: cfg.Section("mongodb").Key("username").String(),
		Password: cfg.Section("mongodb").Key("password").String(),
	}
	// fmt.Println(credential)
	// fmt.Println("MongoDB URI:", options.Client().ApplyURI(uri).SetAuth(credential))
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	mcon.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri).SetAuth(credential))
	if err != nil {
		log.Fatal(os.Stderr, "Invalid MongoDB URI\n")
		panic(err)
	} else {
		fmt.Println("Connected to MongoDB")
	}
	mcon.db = mcon.client.Database(db)
	mcon.colls = make(map[string]*mongo.Collection)
	// fmt.Println(cursor)
}

func (mcon *MongoConnector) disconnect() {
	mcon.client.Disconnect(context.TODO())
}

func (mcon *MongoConnector) insertUser(username string, password string) {
	collection := mcon.getCollection("users")
	_, err := collection.InsertOne(context.TODO(), bson.M{"username": username, "password": password})
	mcon.dbAssert(err != nil, "Error in inserting user", err)
}

func (mcon *MongoConnector) getUser(username string) userRecord {
	collection := mcon.getCollection("users")
	var result bson.M
	err := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&result)
	if mcon.dbAssert(err != nil, "Error in getting user", err) {
		return userRecord{}
	}
	var userdata userRecord = getUserRecord(result)
	fmt.Println(result)
	return userdata
}

func (mcon *MongoConnector) deleteUser(username string) {
	collection := mcon.getCollection("users")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"username": username})
	mcon.dbAssert(err != nil, "Error in deleting user", err)
}

/* CRUD operations for file structure */
func (mcon *MongoConnector) insertFile(folderPath string, fileName string, fileSize int, serverAddr string) string {
	collection := mcon.getCollection("files")
	opts := options.Update().SetUpsert(true)
	inserted_id, err := collection.UpdateOne(context.TODO(), bson.M{
		"folderPath": folderPath, "fileName": fileName, "fileSize": fileSize},
		bson.M{
			"$set": bson.M{
				"lastModified": primitive.NewDateTimeFromTime(time.Now()),
				"serverAddr":   serverAddr,
				"done":         false,
			}}, opts)
	// fmt.Println("Result:", inserted_id)
	mcon.dbAssert(err != nil, "Error in inserting file", err)
	return inserted_id.UpsertedID.(primitive.ObjectID).Hex()
	// _id.(primitive.ObjectID).Hex()
}

func (mcon *MongoConnector) getFile(folderPath string, fileName string) (fileRecord, string) {
	collection := mcon.getCollection("files")
	var result bson.M
	query := bson.M{"folderPath": folderPath, "fileName": fileName}
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	if mcon.dbAssert(err != nil, "Error in getting file", err) {
		return fileRecord{}, ""
	}
	filedata, nodeAddr := getFileRecord(result)
	fmt.Println(filedata)
	return filedata, nodeAddr
}

func (mcon *MongoConnector) updateFileSize(fileId string, fileSize int) int {
	collection := mcon.getCollection("files")
	id, err := primitive.ObjectIDFromHex(fileId)
	updated_id, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id},
		bson.M{"$set": bson.M{"fileSize": fileSize}})
	mcon.dbAssert(err != nil, "Error in updating file size", err)
	return int(updated_id.ModifiedCount)
}

func (mcon *MongoConnector) getFilesFromFolder(folderPath string) []fileRecord {
	collection := mcon.getCollection("files")
	// var result interface{}

	query := bson.M{"folderPath": folderPath}

	cur, err := collection.Find(context.TODO(), query)
	if mcon.dbAssert(err != nil, "Error in getting file from folder", err) {
		return []fileRecord{}
	}
	var filedata []fileRecord
	var result bson.M
	for cur.Next(context.TODO()) {
		cur.Decode(&result)
		file, _ := getFileRecord(result)
		filedata = append(filedata, file)
	}
	return filedata
}

func (mcon *MongoConnector) getFoldersFromFolder(parentFolder string) []folderRecord {
	collection := mcon.getCollection("folders")
	// var result interface{}

	query := bson.M{"parentFolder": parentFolder}

	cur, err := collection.Find(context.TODO(), query)
	if mcon.dbAssert(err != nil, "Error in getting folder from folder", err) {
		return []folderRecord{}
	}
	var folderdata []folderRecord
	var result bson.M
	for cur.Next(context.TODO()) {
		cur.Decode(&result)
		folderdata = append(folderdata, getFolderRecord(result))
	}

	fmt.Println(folderdata)
	return folderdata
}

func (mcon *MongoConnector) deleteFile(folderPath string, fileName string) int {
	collection := mcon.getCollection("files")
	deleted_id, err := collection.DeleteOne(context.TODO(), bson.M{"folderPath": folderPath, "fileName": fileName})
	mcon.dbAssert(err != nil, "Error in deleting file", err)
	return int(deleted_id.DeletedCount)
}

/* CRUD Folder operations */
func (mcon *MongoConnector) insertFolder(parentFolder string, folderName string) (string, time.Time) {
	collection := mcon.getCollection("folders")
	created := time.Now()
	folder := mcon.getFolder(parentFolder, folderName)
	if folder.folderPath != "" {
		return "", created
	}
	inserted_id, err := collection.InsertOne(context.TODO(), bson.M{
		"parentFolder": parentFolder,
		"folderName":   folderName,
		"lastModified": primitive.NewDateTimeFromTime(created),
	})
	mcon.dbAssert(err != nil, "Error in inserting file", err)
	return inserted_id.InsertedID.(primitive.ObjectID).Hex(), created
}

func (mcon *MongoConnector) getFolder(parentFolder string, folderName string) folderRecord {
	collection := mcon.getCollection("folders")
	var result bson.M
	query := bson.M{"parentFolder": parentFolder, "folderName": folderName}
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	if mcon.dbAssert(err != nil, "Error in getting folder", err) {
		return folderRecord{folderPath: ""}
	}
	var folderdata = getFolderRecord(result)
	return folderdata
}

func (mcon *MongoConnector) deleteFolder(parentFolder string, folderName string) int {
	collection := mcon.getCollection("folders")
	deleted_id, err := collection.DeleteOne(context.TODO(), bson.M{"parentFolder": parentFolder, "folderName": folderName})
	mcon.dbAssert(err != nil, "Error in deleting folder", err)
	return int(deleted_id.DeletedCount)
}

func (mcon *MongoConnector) dbAssert(condition bool, message string, err error) bool {
	if condition {
		log.Println(os.Stderr, message, err)
		// panic(message)
		return true
	}
	return false
}

// func (mcon *MongoConnector) addDataNode(serverAddr string) {
// 	collection := mcon.getCollection("servers")
// 	_, err := collection.InsertOne(context.TODO(), bson.M{
// 		"serverAddr": serverAddr,
// 	})
// 	mcon.dbAssert(err != nil, "Error in adding server", err)
// }

func (mcon *MongoConnector) getServers() []DataNode {
	collection := mcon.getCollection("servers")
	cur, err := collection.Find(context.TODO(), bson.M{})
	if mcon.dbAssert(err != nil, "Error in getting servers", err) {
		return []DataNode{}
	}
	var result bson.M
	var dnodes []DataNode
	for cur.Next(context.TODO()) {
		cur.Decode(&result)
		dnodes = append(dnodes, getServer(result))
	}
	return dnodes
}

func (mcon *MongoConnector) updateDataNode(serverAddr string, status bool) int {
	collection := mcon.getCollection("servers")
	opts := options.Update().SetUpsert(true)
	upid, err := collection.UpdateOne(context.TODO(), bson.M{"serverAddr": serverAddr},
		bson.M{"$set": bson.M{"time": primitive.NewDateTimeFromTime(time.Now())}}, opts)
	mcon.dbAssert(err != nil, "Error in updating server", err)
	return int(upid.ModifiedCount)
}

func (mcon *MongoConnector) updateFileDone(fileId string, fileSize int64) int {
	collection := mcon.getCollection("files")
	id, err := primitive.ObjectIDFromHex(fileId)
	updated_id, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"fileSize": fileSize, "lastModified": primitive.NewDateTimeFromTime(time.Now()),
			"done": true,
		}})
	mcon.dbAssert(err != nil, "Error in updating file size", err)
	return int(updated_id.ModifiedCount)
}
