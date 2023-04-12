package main

import "fmt"

func (c *Coordinator) InsertFile(args *InsertFileArgs, reply *InsertFileReply) error {
	// c.mcon.insertFile(args.FilePath, args.FileName)
	fmt.Println("InsertFileReq")
	result, _ := c.mcon.getFile(args.FolderPath, args.FileName)
	if result.id != "" {
		reply.FileId = "File already exists"
		return nil
	}
	reply.FileId = c.mcon.insertFile(args.FolderPath, args.FileName, args.FileSize)
	return nil
}

func (c *Coordinator) DeleteFile(args *DeleteFileArgs, reply *DeleteFileReply) error {
	fmt.Println("DeleteFileReq")
	reply.DeleteCount = c.mcon.deleteFile(args.FolderPath, args.FileName)
	return nil
}

func (c *Coordinator) GetFile(args *GetFileArgs, reply *GetFileReply) error {
	var result fileRecord
	fmt.Println("GetFile")
	result, reply.NodeAddr = c.mcon.getFile(args.FolderPath, args.FileName)
	reply.AccessToken = "AccessToken"
	if result.id == "" {
		reply.FileId = "File doesn't exist"
	}
	reply.FileId = result.id
	return nil
}

func (c *Coordinator) ListFiles(args *ListFilesArgs, reply *ListFilesReply) error {
	// c.mcon.getFilesFromFolder("/home/test1/Desktop")
	// c.mcon.getFile("/home/test1/Desktop", "test1.txt")
	// fmt.Println("ListFilesReq:", args)
	results := c.mcon.getFilesFromFolder(args.FolderPath)
	for _, file := range results {
		storefile := FileStruct{}
		storefile.FileId = file.id
		storefile.FileNames = file.fileName
		storefile.FileModified = file.lastModified
		storefile.FileSize = file.fileSize
		reply.Files = append(reply.Files, storefile)
	}
	return nil
}

func (c *Coordinator) InsertFolder(args *InsertFolderArgs, reply *InsertFolderReply) error {
	fmt.Println("InsertFolderReq")
	result, _ := c.mcon.getFolder(args.ParentPath, args.FolderName)
	return nil
}
