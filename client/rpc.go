package main


/*
 * All the structures created are as per master node
 */
import (
	"net/rpc"
	"time"
)

type User struct {
	name string
	client rpc.Client
	UserToken string
}

func (user *User)connectMaster() bool {
	var err error
	rpc.client, err = rpc.DialHTTP("tcp", "localhost:9000")
	if (err == nil) {
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
func (user *User)getDir() []File {
	args = ListFilesArgs {
		UserToken: user.name,
		FolderPath: "/home/test1/Desktop"		
	}
	user.Call("Coordinator.ListFiles", &args, &reply)
	return reply.FileStruct
}