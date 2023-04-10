package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type userRecord struct {
	username string
	password string
}

type filesRecord struct {
	folderPath   string
	fileName     string
	id           primitive.ObjectID
	lastModified primitive.DateTime
}
