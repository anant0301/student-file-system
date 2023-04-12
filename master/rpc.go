package main

import (
	"time"
)

type ListFilesArgs struct {
	UserToken  string
	FolderPath string
}

type FileStruct struct {
	FileId       string
	FileNames    string
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
	FileId     string
}

type GetFileReply struct {
	NodeAddr    string
	AccessToken string
	FileId      string
}

type InsertFolderArgs struct {
	UserToken  string
	FolderName string
	ParentPath string
}

type InsertFolderReply struct {
	FolderId string
}
