package main

import (
	"context"
	"fmt"
	"os"

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
	mcon.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri).SetAuth(credential))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid MongoDB URI\n")
		panic(err)
	}
	var colls []string
	var nameonly bool = true
	count, err := mcon.client.Database(db).Collection("users").CountDocuments(context.TODO(), bson.D{})
	colls, err = mcon.client.Database(db).ListCollectionNames(context.TODO(), options.ListCollectionsOptions{NameOnly: &nameonly})
	if err != nil {
		fmt.Fprintf(os.Stderr, "MongoDB: Invalid access\n")
		panic(err)
	}
	for _, coll := range colls {
		fmt.Println(coll)
	}
	fmt.Println(count)
}
