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
	FolderId       string
	FolderModified time.Time
}

type WriteFileArgs struct {
	UserToken  string
	FolderPath string
	FileName   string
	FileId     string
}

type WriteFileReply struct {
	ServerId   string
	ServerAddr string
}

type UpdateFileSizeArgs struct {
	FileId   string
	FileSize int
}

type UpdateFileSizeReply struct {
	Done int
}

type PingArgs struct {
	Addr      string
	FreeSpace int
}

type PingReply struct {
	Success bool
	Id      int
}
