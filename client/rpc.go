package main

import (
	"fmt"
	"net/rpc"
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

type InsertFileArgs_c struct {
	Data   []byte
	FileId string
	Offset int64
}

type InsertFileReply_c struct {
	Status   bool
	FileSize int64
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

type DeleteFolderReply struct {
	DeleteCount int
}

type DeleteFolderArgs struct {
	UserToken  string
	FolderName string
	ParentPath string
}

type InsertFolderReply struct {
	FolderId       string
	FolderModified time.Time
}

type PingArgs struct {
	Addr      string
	FreeSpace int
}

type PingReply struct {
	Success bool
	Id      int
}

type CreateFileArgs_dn struct {
	FileId string
}

type CreateFileReply_dn struct {
	Status bool
}

type CreateFileArgs_m struct {
	UserToken  string
	FolderPath string
	FileName   string
}

type CreateFileReply_m struct {
	FileId     string
	ServerAddr string
}

type GetFileArgs_c struct {
	AccessToken string
	FileId      string
	Offset      int64
	SizeOfChunk int
}

type GetFileReply_c struct {
	Status bool
	Data   []byte
}

type RenameFileArgs struct {
	OldPath string
	NewPath string
	OldName string
	NewName string
}

type RenameFileReply struct {
	Status bool
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func Call(rpcname string, args interface{}, reply interface{}) error {

	c, err := rpc.DialHTTP("tcp", "10.7.50.133"+":9000")

	if err != nil {
		fmt.Println("Can't connect to server: ", err)
		return err
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)

	return err
}

func CallDataNode(addr string, rpcname string, args interface{}, reply interface{}) error {

	c, err := rpc.DialHTTP("tcp", addr)

	if err != nil {
		fmt.Println("Can't connect to server: ", addr, err)
		return err
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)

	return err
}