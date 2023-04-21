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

func getFileRecord(file bson.M) (fileRecord, string) {
	var filedata fileRecord
	filedata.folderPath = file["folderPath"].(string)
	filedata.fileName = file["fileName"].(string)
	filedata.id = file["_id"].(primitive.ObjectID).Hex()
	filedata.lastModified = file["lastModified"].(primitive.DateTime).Time()
	filedata.fileSize = 100
	return filedata, file["serverAddr"].(string)
}

func getFolderRecord(folder bson.M) folderRecord {
	var folderdata folderRecord
	folderdata.folderPath = folder["folderName"].(string)
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

func getServer(server bson.M) DataNode {
	var dnode DataNode
	lastModified := server["time"].(primitive.DateTime).Time()
	if time.Since(lastModified) > 60*time.Second {
		dnode.IsAlive = false
	} else {
		dnode.IsAlive = true
	}
	dnode.Addr = server["serverAddr"].(string)
	return dnode
}
