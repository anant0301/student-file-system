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
}

type InsertFileReply struct {
	NodeAddr    string
	IsLocked    bool
	AccessToken string
	File        FileStruct
}

type DeleteFileArgs struct {
	UserToken  string
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

type CreateFileArgs struct {
	UserToken  string
	FolderPath string
	FileName   string
}

type CreateFileReply struct {
	FileId     string
	ServerAddr string
}

type RenameFileArgs struct {
	UserToken string
	OldPath   string
	NewPath   string
	OldName   string
	NewName   string
}

type RenameFileReply struct {
	Status bool
}

type DeleteFolderArgs struct {
	UserToken  string
	ParentPath string
	FolderName string
}

type DeleteFolderReply struct {
	DeleteCount int
}

type UpdateFileSizeArgs struct {
	FileId   string
	FileSize int
}

type UpdateFileSizeReply struct {
	Done int
}
