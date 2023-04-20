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
		var folder folderRecord
		folder, reply.NodeAddr = c.mcon.getFolder(args.FolderPath, args.FileName)
		if folder.folderId == "" {
			reply.File.FileId = "0" // File not found
			return nil
		}
		reply.File.FileId = folder.folderId
		reply.File.FileName = folder.folderPath
		reply.File.FileModified = folder.lastModified
		reply.File.FileSize = 0
		reply.File.IsFolder = true
	} else {
		reply.File.FileId = result.id
		reply.File.FileName = result.fileName
		reply.File.FileModified = result.lastModified
		reply.File.FileSize = result.fileSize
		reply.File.IsFolder = false
	}
	return nil
}

func (c *Coordinator) UpdateFileSize(args *UpdateFileSizeArgs, reply *UpdateFileSizeReply) error {
	fmt.Println("UpdateFileSizeReq")
	reply.Done = c.mcon.updateFileSize(args.FileId, args.FileSize)
	return nil
}

func (c *Coordinator) WriteFile(args *WriteFileArgs, reply *WriteFileReply) error {
	fmt.Println("WriteFileReq")
	return nil
}

func (c *Coordinator) ListFiles(args *ListFilesArgs, reply *ListFilesReply) error {
	// c.mcon.getFilesFromFolder("/home/test1/Desktop")
	// c.mcon.getFile("/home/test1/Desktop", "test1.txt")
	fmt.Println("ListFilesReq:", args)
	results := c.mcon.getFilesFromFolder(args.FolderPath)
	for _, file := range results {
		storefile := FileStruct{}
		storefile.FileId = file.id
		storefile.FileName = file.fileName
		storefile.FileModified = file.lastModified
		storefile.FileSize = file.fileSize
		storefile.IsFolder = false
		reply.Files = append(reply.Files, storefile)
	}
	folders := c.mcon.getFoldersFromFolder(args.FolderPath)
	for _, folder := range folders {
		storefile := FileStruct{}
		storefile.FileId = folder.folderId
		storefile.FileName = folder.folderPath
		storefile.FileModified = folder.lastModified
		storefile.FileSize = 0
		storefile.IsFolder = true
		reply.Files = append(reply.Files, storefile)
	}
	return nil
}

func (c *Coordinator) InsertFolder(args *InsertFolderArgs, reply *InsertFolderReply) error {
	fmt.Println("InsertFolderReq")
	reply.FolderId, reply.FolderModified = c.mcon.insertFolder(args.ParentPath, args.FolderName)
	if reply.FolderId == "" {
		reply.FolderId = "0"
	}
	return nil
}

// func (c *Coordinator) DeleteFolder(args *DeleteFolderArgs, reply *DeleteFolderReply) error {
// 	fmt.Println("DeleteFolderReq")
// 	reply.DeleteCount = c.mcon.deleteFolder(args.FolderPath, args.FolderName)
// 	return nil
// }
