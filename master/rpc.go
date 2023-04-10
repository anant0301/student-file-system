package main

import "time"

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
