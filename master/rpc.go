package main

import (
	"fmt"
	"time"
)

type ListFilesArgs struct {
	UserToken  string
	FolderPath string
}

type ListFilesReply struct {
	FileId       []int
	FileNames    []string
	IsFolder     []bool
	FileModified []time.Time
	FileSize     []int
}

type InsertFileArgs struct {
	UserToken  string
	FolderPath string
	FileName   string
	FileSize   int
}

type InsertFileReply struct {
	FileId int
}

func (c *Coordinator) InsertFileReq(args *InsertFileArgs, reply *InsertFileReply) {
	// c.mcon.insertFile(args.FilePath, args.FileName)
	c.mcon.insertFile("/home/test1/Desktop", "test1.txt")
	c.mcon.insertFile("/home/test1/Desktop", "test2.txt")
	fmt.Println("InsertFileReq")
}

func (c *Coordinator) ListFilesReq(args *ListFilesArgs, reply *ListFilesReply) {
	// c.mcon.getFilesFromFolder("/home/test1/Desktop")
	c.mcon.getFile("/home/test1/Desktop", "test1.txt")
	fmt.Println("ListFilesReq")
}
