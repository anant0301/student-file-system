package main

import (
	"time"
)

type FileStruct struct {
	FileId       string
	FileName     string
	IsFolder     bool
	FileModified time.Time
	FileSize     int64
}

type InsertFileDoneArgs struct {
	BytesWritten int64
}
type InsertFileDoneReply struct {
	Status bool
}

type GetReplicationNodesArgs struct {
	//plis send nodes
}
type GetReplicationNodesReply struct {
	ReplicationNodes []string
	//Status bool
}

type InsertFileArgs_c struct {
	Data         []byte
	FileId       string
	Offset       int64
}

type InsertFileReply_c struct {
	Status bool
	FileSize	int64
}

type InsertFileArgs_m struct {
	UserToken string
	//ClientAddr string
}

type InsertFileReply_m struct {
	Status bool
}

type CreateFileArgs_c struct {
	//Data   string
	FileId string
	//Offset int64
}

type CreateFileReply_c struct {
	Status bool
}

type CreateFileArgs_m struct {
	UserToken string
	//ClientAddr string
}

type CreateFileReply_m struct {
	Status bool
}

// type DeleteFileArgs_c struct {
// 	AccessToken string
// 	FileId      string
// }

// type DeleteFileReply_c struct {
// 	Status bool
// }

// type DeleteFileArgs_m struct {
// 	AccessToken      string
// 	ClientAddr       string
// 	ReplicationNodes []string
// }

// type DeleteFileReply_m struct {
// 	Status bool
// }

type GetFileArgs_m struct {
	AccessToken string
	ClientAddr  string
}

type GetFileReply_m struct {
	Status bool
	Data   []byte
}

type GetFileArgs_c struct {
	AccessToken string
	FileId      string
	Offset      int64
	SizeOfChunk int64
}
type GetFileReply_c struct {
	Status bool
	Data   []byte
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
	//Id        int
}

type PingReply struct {
	Status bool
	//Id     int
}

type DoneArgs struct {
	FileId    string
	Operation string
	FileSize  int64
}

type DoneReply struct {
	Status bool
}

type DeleteFileArgs_m struct{
	FileId string
	ReplicationNodes []string
	UserToken	string
}

type DeleteFileReply_m struct{
	Status bool
}
