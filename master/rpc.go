package main

import "time"

type ListFileArgs struct {
	UserToken  string
	FolderPath string
}

type ListFileReply struct {
	FileId       []int
	FileNames    []string
	IsFolder     []bool
	FileModified []time.Time
}
