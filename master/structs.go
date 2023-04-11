package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userRecord struct {
	username string
	password string
}

type filesRecord struct {
	folderPath   string
	fileName     string
	id           string
	lastModified time.Time
}

func getFileRecord(file bson.M) filesRecord {
	var filedata filesRecord
	filedata.folderPath = file["folderPath"].(string)
	filedata.fileName = file["fileName"].(string)
	filedata.id = file["_id"].(primitive.ObjectID).Hex()
	filedata.lastModified = file["lastModified"].(primitive.DateTime).Time()
	return filedata
}

func getUserRecord(user bson.M) userRecord {
	var userdata userRecord
	userdata.username = user["username"].(string)
	userdata.password = user["password"].(string)
	return userdata
}
