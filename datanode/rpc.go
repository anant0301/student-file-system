package main

import (
	"time"
)

type FileStruct struct {
	FileId       string
	FileName     string
	IsFolder     bool
	FileModified time.Time
	FileSize     int
}

type InsertFileArgs_c struct {
	Data   string
	FileId string
}

type InsertFileReply_c struct {
	Status bool
	BytesWritten []byte
}

type InsertFileArgs_m struct {
	AccessToken      string
	ClientAddr       string
	ReplicationNodes []string
}

type InsertFileReply_m struct {
	Status bool
}

type DeleteFileArgs_c struct {
	AccessToken      string
	FileId      string
}

type DeleteFileReply_c struct {
	Status	bool
}

type DeleteFileArgs_m struct {
	AccessToken      string
	ClientAddr       string
	ReplicationNodes []string
}

type DeleteFileReply_m struct {
	Status	bool
}

type GetFileArgs_m struct {
	AccessToken string
	ClientAddr  string
}

type GetFileReply_m struct {
	Status bool
}

type GetFileArgs_c struct {
	AccessToken string
	FileId      string
}
type GetFileReply_c struct {
	Status bool
}

type DataNodeData struct {
	Data string
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
