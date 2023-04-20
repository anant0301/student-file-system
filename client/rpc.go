package main

import (
	"fmt"
	"net/rpc"
	"os"
	"time"
)

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

type InsertFileArgs struct {
	UserToken  string
	FolderPath string
	FileName   string
	FileSize   int
}

type InsertFileReply struct {
	FileId string
}

type DeleteFileArgs struct {
	UserToken  string
	FileId     string
	FolderPath string
	FileName   string
}

type DeleteFileReply struct {
	DeleteCount int
}

type GetFileArgs struct {
	UserToken  string
	FolderPath string
	FileName   string
}

type GetFileReply struct {
	NodeAddr    string
	AccessToken string
	File        FileStruct
}

type InsertFolderArgs struct {
	UserToken  string
	FolderName string
	ParentPath string
}

type InsertFolderReply struct {
	FolderId string
}

type PingArgs struct {
	Addr      string
	FreeSpace int
}

type PingReply struct {
	Success bool
	Id      int
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func Call(rpcname string, args interface{}, reply interface{}) error {

	c, err := rpc.DialHTTP("tcp", "10.7.50.133"+":9000")

	if err != nil {
		fmt.Println("Can't connect to server. Exiting....")
		os.Exit(0)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)

	return err
}
