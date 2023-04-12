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
		fmt.Fprintf(os.Stderr, "Invalid MongoDB URI\n")
		panic(err)
	}
	mcon.db = mcon.client.Database(db)
	mcon.colls = make(map[string]*mongo.Collection)
	// fmt.Println(cursor)
	fmt.Println("Connected to MongoDB")
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
func (mcon *MongoConnector) insertFile(folderPath string, fileName string, fileSize int) string {
	collection := mcon.getCollection("files")
	inserted_id, err := collection.InsertOne(context.TODO(), bson.M{
		"folderPath": folderPath, "fileName": fileName, "fileSize": fileSize,
		"lastModified": primitive.NewDateTimeFromTime(time.Now()),
	})
	mcon.dbAssert(err != nil, "Error in inserting file", err)
	return inserted_id.InsertedID.(primitive.ObjectID).Hex()
}

func (mcon *MongoConnector) getFile(folderPath string, fileName string) (fileRecord, string) {
	collection := mcon.getCollection("files")
	var result bson.M
	query := bson.M{"folderPath": folderPath, "fileName": fileName}
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	nodeAddr := getFileNode("filename")
	if mcon.dbAssert(err != nil, "Error in getting file", err) {
		return fileRecord{}, nodeAddr
	}
	var filedata = getFileRecord(result)
	fmt.Println(filedata)
	return filedata, nodeAddr
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
		filedata = append(filedata, getFileRecord(result))
	}

	fmt.Println(filedata)
	return filedata
}

func (mcon *MongoConnector) deleteFile(folderPath string, fileName string) int {
	collection := mcon.getCollection("files")
	deleted_id, err := collection.DeleteOne(context.TODO(), bson.M{"folderPath": folderPath, "fileName": fileName})
	mcon.dbAssert(err != nil, "Error in deleting file", err)
	return int(deleted_id.DeletedCount)
}

/* CRUD Folder operations */
func (mcon *MongoConnector) insertFolder(parentFolder string, folderName string) string {
	collection := mcon.getCollection("folders")
	inserted_id, err := collection.InsertOne(context.TODO(), bson.M{
		"parentFolder": parentFolder,
		"folderName":   folderName,
		"lastModified": primitive.NewDateTimeFromTime(time.Now()),
	})
	mcon.dbAssert(err != nil, "Error in inserting file", err)
	return inserted_id.InsertedID.(primitive.ObjectID).Hex()
}

func (mcon *MongoConnector) getFolder(parentFolder string, folderName string) (folderRecord, string) {
	collection := mcon.getCollection("folders")
	var result bson.M
	query := bson.M{"parentFolder": parentFolder, "folderName": folderName}
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	nodeAddr := getFileNode("foldername")
	if mcon.dbAssert(err != nil, "Error in getting folder", err) {
		return folderRecord{}, nodeAddr
	}
	var folderdata = getFolderRecord(result)
	fmt.Println(folderdata)
	return folderdata, nodeAddr
}

func (mcon *MongoConnector) dbAssert(condition bool, message string, err error) bool {
	if condition {
		log.Println(os.Stderr, message, err)
		// panic(message)
		return true
	}
	return false
}