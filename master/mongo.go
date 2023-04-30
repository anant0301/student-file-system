package main

import (
	"context"
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
		log.Println("Error in parsing .ini file for mongo credentials\n")
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
	// log.Println(credential)
	// log.Println("MongoDB URI:", options.Client().ApplyURI(uri).SetAuth(credential))
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	mcon.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri).SetAuth(credential))
	if err != nil {
		log.Fatal(os.Stderr, "Invalid MongoDB URI\n")
		panic(err)
	} else {
		log.Println("Connected to MongoDB")
	}
	mcon.db = mcon.client.Database(db)
	mcon.colls = make(map[string]*mongo.Collection)
	// log.Println(cursor)
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
	log.Println(result)
	return userdata
}

func (mcon *MongoConnector) deleteUser(username string) {
	collection := mcon.getCollection("users")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"username": username})
	mcon.dbAssert(err != nil, "Error in deleting user", err)
}

func (mcon *MongoConnector) getLock(fileId string) bool {
	collection := mcon.getCollection("files")
	id, err := primitive.ObjectIDFromHex(fileId)
	if mcon.dbAssert(err != nil, "Error converting string to ObjectId", err) {
		return false
	}
	upsert := false
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	var result bson.M
	err = collection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id, "done": true}, bson.M{"$set": bson.M{"done": false}}, &opt).Decode(&result)
	if mcon.dbAssert(err != nil, "Error in releasing lock", err) {
		return false
	}
	return true
}

func (mcon *MongoConnector) releaseLock(fileId string) bool {
	collection := mcon.getCollection("files")
	id, err := primitive.ObjectIDFromHex(fileId)
	if mcon.dbAssert(err != nil, "Error converting string to ObjectId", err) {
		return false
	}
	upsert := false
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	var result bson.M
	err = collection.FindOneAndUpdate(context.TODO(), bson.M{"_id": id, "done": false}, bson.M{"$set": bson.M{"done": true}}, &opt).Decode(&result)
	if mcon.dbAssert(err != nil, "Error in releasing lock", err) {
		return false
	}
	return true
}

/* CRUD operations for file structure */
func (mcon *MongoConnector) insertFile(folderPath string, fileName string, fileSize int) string {
	collection := mcon.getCollection("files")
	log.Println("Inserting file:", folderPath, fileName)
	inserted_id, err := collection.InsertOne(context.TODO(), bson.M{
		"folderPath": folderPath, "fileName": fileName, "fileSize": fileSize,
		"lastModified": primitive.NewDateTimeFromTime(time.Now()),
		"done":         false,
	})
	// log.Println("Result:", inserted_id)
	mcon.dbAssert(err != nil, "Error in inserting file", err)
	return inserted_id.InsertedID.(primitive.ObjectID).Hex()
	// _id.(primitive.ObjectID).Hex()
}

func (mcon *MongoConnector) getFile(folderPath string, fileName string) fileRecord {
	collection := mcon.getCollection("files")
	var result bson.M
	log.Println("Get file:", folderPath, fileName)
	query := bson.M{"folderPath": folderPath, "fileName": fileName}
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	if mcon.dbAssert(err != nil, "Error in getting file", err) {
		return fileRecord{}
	}
	filedata := getFileRecord(result)
	logCollection := mcon.getCollection("logs")
	logCollection.FindOne(context.TODO(), bson.M{"fileId": filedata.id}).Decode(&result)
	filedata.lastModified = result["lastModified"].(primitive.DateTime).Time()
	log.Println(filedata)
	return filedata
}

func (mcon *MongoConnector) updateFileSize(fileId string, fileSize int) int {
	collection := mcon.getCollection("files")
	id, err := primitive.ObjectIDFromHex(fileId)
	mcon.dbAssert(err != nil, "Error in creating bson", err)
	updated_id, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id},
		bson.M{"$set": bson.M{"$fileSize": fileSize}})
	mcon.dbAssert(err != nil, "Error in updating file size", err)
	return int(updated_id.ModifiedCount)
}

func (mcon *MongoConnector) getFileSize(fileId string) int {
	collection := mcon.getCollection("files")
	id, err := primitive.ObjectIDFromHex(fileId)
	mcon.dbAssert(err != nil, "Error in creating bson", err)
	var result bson.M
	err = collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)
	mcon.dbAssert(err != nil, "Error in getting file size", err)
	log.Println("GetFileSize", result, fileId)
	return int(result["fileSize"].(int32))
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
		file := getFileRecord(result)
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
	return folderdata
}

func (mcon *MongoConnector) deleteFile(folderPath string, fileName string) int {
	collection := mcon.getCollection("files")
	log.Println("Deleting file:", folderPath, fileName)
	deleted_id, err := collection.DeleteOne(context.TODO(), bson.M{"folderPath": folderPath, "fileName": fileName})
	mcon.dbAssert(err != nil, "Error in deleting file", err)
	return int(deleted_id.DeletedCount)
}

func (mcon *MongoConnector) renameFile(oldPath string, oldName string, newPath string, newName string) bool {
	collection := mcon.getCollection("files")
	collection.DeleteOne(context.TODO(), bson.M{"folderPath": newPath, "fileName": newName})
	updated_id, err := collection.UpdateOne(context.TODO(), bson.M{"folderPath": oldPath, "fileName": oldName},
		bson.M{"$set": bson.M{"folderPath": newPath, "fileName": newName}})
	if mcon.dbAssert(err != nil, "Error in renaming file", err) {
		return false
	}
	return updated_id.ModifiedCount == 1
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

func (mcon *MongoConnector) renameFolder(oldPath string, oldName string, newPath string, newName string) bool {
	collection := mcon.getCollection("folders")

	updated_id, err := collection.UpdateOne(context.TODO(), bson.M{"parentFolder": oldPath, "folderName": oldName},
		bson.M{"$set": bson.M{"parentFolder": newPath, "folderName": newName}})
	if mcon.dbAssert(err != nil, "Error in renaming folder", err) {
		return false
	}
	return updated_id.ModifiedCount == 1
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

/* Data node related queries */
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
		res := getServer(result)
		if res.IsAlive {
			dnodes = append(dnodes, res)
		}
	}
	return dnodes
}

func (mcon *MongoConnector) updateDataNode(serverAddr string) int {
	collection := mcon.getCollection("servers")
	opts := options.Update().SetUpsert(true)
	upid, err := collection.UpdateOne(context.TODO(), bson.M{"serverAddr": serverAddr},
		bson.M{"$set": bson.M{"time": primitive.NewDateTimeFromTime(time.Now())}}, opts)
	mcon.dbAssert(err != nil, "Error in updating server", err)
	return int(upid.ModifiedCount)
}

func (mcon *MongoConnector) updateFileDone(fileId string, fileSize int, doneTime time.Time) int {
	collection := mcon.getCollection("files")
	id, err := primitive.ObjectIDFromHex(fileId)
	updated_id, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"fileSize": fileSize, "lastModified": primitive.NewDateTimeFromTime(doneTime),
			"done": true,
		}})
	mcon.dbAssert(err != nil, "Error in updating file size", err)
	return int(updated_id.ModifiedCount)
}

func (mcon *MongoConnector) updatePing(serverId string) int {
	collection := mcon.getCollection("servers")
	opts := options.Update().SetUpsert(true)
	updated_id, err := collection.UpdateOne(context.TODO(), bson.M{"serverAddr": serverId},
		bson.M{"$set": bson.M{"time": primitive.NewDateTimeFromTime(time.Now())}}, opts)
	mcon.dbAssert(err != nil, "Error in updating server last operation", err)
	return int(updated_id.ModifiedCount)
}

/* log replication related query */
func (mcon *MongoConnector) updateLogsNode(serverId string, fileId string, operation string, doneTime time.Time) string {
	collection := mcon.getCollection("logs")
	opts := options.Update().SetUpsert(true)
	updated_id, err := collection.UpdateOne(context.TODO(), bson.M{"fileId": fileId, "serverId": serverId},
		bson.M{"$set": bson.M{"operation": operation,
			"lastUpdated": primitive.NewDateTimeFromTime(doneTime)}}, opts)
	mcon.dbAssert(err != nil, "Error in updating logs", err)
	if updated_id.UpsertedID != nil {
		return updated_id.UpsertedID.(primitive.ObjectID).Hex()
	}
	return ""
}

func (mcon *MongoConnector) getServerLastOpTime(serverId string) time.Time {
	collection := mcon.getCollection("logs")
	var result bson.M
	opts := options.Find().SetSort(bson.D{{"enrollment", -1}})
	cur, err := collection.Find(context.TODO(), bson.M{
		"serverId": serverId,
	}, opts)
	if err != nil {
		log.Println("Error in getting logs")
		return time.Time{}
	}
	for cur.Next(context.TODO()) {
		cur.Decode(&result)
		break
	}
	log.Println("resutls", result)
	if result != nil {
		return result["lastUpdated"].(primitive.DateTime).Time()
	}
	return time.Time{}
}

func (mcon *MongoConnector) getInconsistentLogs(serverId string) []logRecord {
	collection := mcon.getCollection("logs")
	aliveTime := mcon.getServerLastOpTime(serverId)
	// if aliveTime.Sub(time.Time{}) == 0 {
	// 	aliveTime = time.Now().Sin
	// }

	var result []logRecord

	cur, err := collection.Find(
		context.TODO(),
		bson.M{
			"lastUpdated": bson.M{
				"$gt": primitive.NewDateTimeFromTime(aliveTime),
			}})
	if mcon.dbAssert(err != nil, "Error in getting logs", err) {
		log.Println("Error in getting logs")
		return []logRecord{}
	}

	for cur.Next(context.TODO()) {
		var elem bson.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		if elem["serverId"] == serverId {
			continue
		}
		log.Println("elem", elem)
		logdata := getLogRecord(elem)
		if logdata.lastUpdated.Sub(aliveTime) < 0 {
			continue
		}
		if logdata.operation != "delete" {
			logdata.fileSize = mcon.getFileSize(logdata.fileId)
		}
		result = append(result, logdata)
	}
	return result
}
