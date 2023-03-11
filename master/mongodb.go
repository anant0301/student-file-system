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
	coll   *mongo.Collection
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
	mcon.client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri).SetAuth(credential))
	defer mcon.client.Disconnect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid MongoDB URI\n")
		panic(err)
	}

	// var colls []string
	var nameonly bool = true
	cursor, err := mcon.client.Database(db).Collection("users").Find(ctx, bson.D{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "MongoDB: Invalid access\n")
		panic(err)
	}
	colls, err := mcon.client.Database(db).ListCollectionNames(ctx, options.ListCollectionsOptions{NameOnly: &nameonly})
	if err != nil {
		fmt.Fprintf(os.Stderr, "MongoDB: Invalid access\n")
		panic(err)
	}
	for _, coll := range colls {
		fmt.Println(coll)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var episode bson.M
		if err = cursor.Decode(&episode); err != nil {
			log.Fatal(err)
		}
		fmt.Println(episode)
	}
	// fmt.Println(cursor)
}
