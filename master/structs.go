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

type fileRecord struct {
	folderPath   string
	fileName     string
	id           string
	lastModified time.Time
	fileSize     int
}

type folderRecord struct {
	folderPath   string
	parentFolder string
	lastModified time.Time
	folderId     string
}

type nodeRecord struct {
	nodeAddr string
	nodeId   string
}

func getFileRecord(file bson.M) fileRecord {
	var filedata fileRecord
	filedata.folderPath = file["folderPath"].(string)
	filedata.fileName = file["fileName"].(string)
	filedata.id = file["_id"].(primitive.ObjectID).Hex()
	filedata.lastModified = file["lastModified"].(primitive.DateTime).Time()
	filedata.fileSize = 100
	return filedata
}

func getFolderRecord(folder bson.M) folderRecord {
	var folderdata folderRecord
	folderdata.folderPath = folder["folderPath"].(string)
	folderdata.parentFolder = folder["parentFolder"].(string)
	folderdata.lastModified = folder["lastModified"].(primitive.DateTime).Time()
	folderdata.folderId = folder["_id"].(primitive.ObjectID).Hex()
	return folderdata
}

func getUserRecord(user bson.M) userRecord {
	var userdata userRecord
	userdata.username = user["username"].(string)
	userdata.password = user["password"].(string)
	return userdata
}

func getFileNode(fileId string) string {
	return "localhost:8080"
}
