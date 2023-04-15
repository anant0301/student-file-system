package main

/*
 * All the structures created are as per master node
 */
import (
	"log"
	"net/rpc"
	"time"
)

type User struct {
	client *rpc.Client
}

func (user *User) connectMaster(host string) bool {
	var err error
	user.client, err = rpc.DialHTTP("tcp", host)
	if err != nil {
		log.Fatal("Error in connecting to the Master node")
		return false
	}
	return true
}

type ListFilesArgs struct {
	UserToken  string
	FolderPath string
}

type FileStruct struct {
	FileId       string
	FileName     string
	IsFolder     bool
	FileModified time.Time
	FileSize     int
}

type ListFilesReply struct {
	Files []FileStruct
}

// A temporary function that sends list of entries in a directory
func getDir() []FileStruct {
	args := ListFilesArgs{
		UserToken:  "test1",
		FolderPath: "/home/test1/Desktop",
	}
	reply := ListFilesReply{}
	user := getUser()
	_ = user.client.Call("Coordinator.ListFiles", &args, &reply)
	// if err != nil {
	// 	log.Fatal("Error in calling ListFiles")
	// }
	return reply.Files
}
