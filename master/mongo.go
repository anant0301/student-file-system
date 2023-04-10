package main

import (
	"context"
	"fmt"
	"os"

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

func (mcon *MongoConnector) getUser(username string) {
	collection := mcon.getCollection("users")
	var result bson.M
	err := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&result)
	mcon.dbAssert(err != nil, "Error in getting user", err)
	var userdata userRecord
	userdata.username = result["username"].(string)
	userdata.password = result["password"].(string)
	fmt.Println(result)
}

func (mcon *MongoConnector) deleteUser(username string) {
	collection := mcon.getCollection("users")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"username": username})
	mcon.dbAssert(err != nil, "Error in deleting user", err)
}

func (mcon *MongoConnector) insertFile(folderPath string, fileName string) bool {
	collection := mcon.getCollection("files")
	_, err := collection.InsertOne(context.TODO(), bson.M{"folderPath": folderPath, "fileName": fileName})
	mcon.dbAssert(err != nil, "Error in inserting file", err)
	return err = nil
}

/* CRUD operations for file structure */

func (mcon *MongoConnector) getFile(folderPath string, fileName string) filesRecord {
	collection := mcon.getCollection("files")
	var result bson.M
	// var result interface{}

	query := bson.M{"folderPath": folderPath, "fileName": fileName}
	if folderPath == "" {
		query = bson.M{"fileName": fileName}
	}
	if fileName == "" {
		query = bson.M{"folderPath": folderPath}
	}
	if folderPath == "" && fileName == "" {
		query = bson.M{}
	}
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	mcon.dbAssert(err != nil, "Error in getting file", err)
	var filedata filesRecord
	filedata.folderPath = result["folderPath"].(string)
	filedata.fileName = result["fileName"].(string)
	filedata.id = result["_id"].(primitive.ObjectID)
	filedata.lastModified = result["lastModified"].(primitive.DateTime)
	fmt.Println(filedata)
	return filedata
}

func (mcon *MongoConnector) getFilesFromFolder(folderPath string) {
	collection := mcon.getCollection("files")
	var result []bson.M
	// var result interface{}

	query := bson.M{"folderPath": folderPath, "fileName": fileName}
	if folderPath == "" {
		query = bson.M{"fileName": fileName}
	}
	if fileName == "" {
		query = bson.M{"folderPath": folderPath}
	}
	if folderPath == "" && fileName == "" {
		query = bson.M{}
	}
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	mcon.dbAssert(err != nil, "Error in getting file", err)
	var filedata []filesRecord
	for _, file := range result {
		filedata.folderPath = append(filedata.folderPath, file["folderPath"].(string))
		filedata.fileName = append(filedata.fileName, file["fileName"].(string))
		filedata.id = append(filedata.id, file["_id"].(primitive.ObjectID))
		filedata.lastModified = append(filedata.lastModified, file["lastModified"].(string))
	}
	fmt.Println(filedata)
	return filedata
}

func (mcon *MongoConnector) deleteFile(folderPath string, fileName string) {
	collection := mcon.getCollection("files")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"folderPath": folderPath, "fileName": fileName})
	mcon.dbAssert(err != nil, "Error in deleting file", err)
}

func (mcon *MongoConnector) dbAssert(condition bool, message string, err error) {
	if condition {
		fmt.Fprintf(os.Stderr, message, err)
		panic(message)
	}
}
