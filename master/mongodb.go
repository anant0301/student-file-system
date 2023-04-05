package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/ini.v1"
	// "github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnector struct {
	client *mongo.Client
	db     *mongo.Database
	// dbName string
}

type userCollection struct {
	username string `bson:"username", json:"username"`
}

type filesCollectionRecord struct {
	folderPath string `bson:"folderPath,omitempty", json:"folderPath,omitempty"`
	fileName   string `bson:"fileName,omitempty", json:"fileName,omitempty"`
	fileId     string `bson:"fileId,omitempty", json:"fileId,omitempty"`
}

func (mcon *MongoConnector) dbAssert(cond bool, errMsg string) {
	if cond {
		fmt.Printf("Error in DB: %s\n", errMsg)
	}
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
		AuthSource:    db,
		AuthMechanism: "SCRAM-SHA-256",
		Username:      cfg.Section("mongodb").Key("username").String(),
		Password:      cfg.Section("mongodb").Key("password").String(),
	}
	// fmt.Println("MongoDB URI:", options.Client().ApplyURI(uri).SetAuth(credential))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	mcon.client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri).SetAuth(credential))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid MongoDB URI\n")
		panic(err)
	}
	mcon.db = mcon.client.Database(db)
	mcon.getUsers("test1")
	// fmt.Println(cursor)
}

// CRUD operations on user collection
func (mcon *MongoConnector) getUsers(username string) {

	// var colls []string
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := mcon.db.Collection("users").Find(ctx, bson.D{{"username", username}})
	mcon.dbAssert(err != nil, "getUser: Error in User Fetch")

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var episode bson.M
		if err = cursor.Decode(&episode); err != nil {
			log.Fatal(err)
		}
		fmt.Println(episode)
	}
}

// CRUD operations on file collection
func (mcon *MongoConnector) getFilesFromFolder(folderPath string) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	log.Println("folderPath: ", folderPath)
	cursor, err := mcon.db.Collection("files").Find(ctx, bson.D{{"folderPath", folderPath}})
	mcon.dbAssert(err != nil, "getUser: Error in User Fetch")

	defer cursor.Close(ctx)
	var results []filesCollectionRecord
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf("%+v\n", result)
	}
	// for cursor.Next(ctx) {
	// 	var episode bson.M
	// 	var filesData = filesCollectionRecord{}
	// 	if err = cursor.Decode(&episode); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	bson.Unmarshal(episode, &filesData)
	// }
}
