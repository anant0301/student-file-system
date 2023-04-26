package main

import (
	"log"
	"time"
)

func (c *Coordinator) InsertFile(args *InsertFileArgs, reply *InsertFileReply) error {
	// c.mcon.insertFile(args.FilePath, args.FileName)

	log.Println("GetFile")
	var result fileRecord
	result = c.mcon.getFile(args.FolderPath, args.FileName)
	reply.AccessToken = "AccessToken"
	reply.File.FileId = result.id
	reply.File.FileName = result.fileName
	reply.File.FileModified = result.lastModified
	reply.File.FileSize = result.fileSize
	reply.File.IsFolder = false

	if !c.mcon.getLock(reply.File.FileId) {
		reply.IsLocked = false
		return nil
	} else {
		reply.IsLocked = true
	}
	dnodes := c.mcon.getServers()
	flag := false
	for _, dnode := range dnodes {
		if dnode.IsAlive == true {
			reply.NodeAddr = dnode.Addr
			flag = true
			break
		}
	}
	if flag {
		c.mcon.releaseLock(result.id)
	}
	return nil
}

func (c *Coordinator) DeleteFile(args *DeleteFileArgs, reply *DeleteFileReply) error {
	log.Println("DeleteFileReq:", args.FolderPath, args.FileName)
	file := c.mcon.getFile(args.FolderPath, args.FileName)
	if file.id == "" {
		reply.DeleteCount = 0
		return nil
	}

	var dnodes []DataNode
	dnodes = c.mcon.getServers()
	var addrs []string
	for _, dnode := range dnodes {
		addrs = append(addrs, dnode.Addr)
	}
	for _, dnode := range dnodes {
		if dnode.IsAlive == true {
			DeleteArgs := DeleteFileArgs_m{UserToken: "Access Token", ReplicationNodes: addrs, FileId: file.id}
			DeleteReply := DeleteFileReply_m{}
			ok := c.DialDataNode(dnode.Addr, "DataNode.DeleteFile_m", &DeleteArgs, &DeleteReply)
			if ok == nil && DeleteReply.Status {
				reply.DeleteCount = c.mcon.deleteFile(args.FolderPath, args.FileName)
				// reply.DeleteCount = c.mcon.deleteLogs(args.FolderPath, args.FileName)
				c.mcon.updateLogsNode(dnode.Addr, file.id, DELETE, time.Now())
				break
			} else {
				reply.DeleteCount = 0
			}
		}
	}
	return nil
}

func (c *Coordinator) GetFile(args *GetFileArgs, reply *GetFileReply) error {
	log.Println("GetFile")
	var result fileRecord
	result = c.mcon.getFile(args.FolderPath, args.FileName)

	reply.AccessToken = "AccessToken"
	if result.id == "" {
		var folder folderRecord
		folder = c.mcon.getFolder(args.FolderPath, args.FileName)
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
	dnodes := c.mcon.getServers()
	for _, dnode := range dnodes {
		if dnode.IsAlive == true {
			reply.NodeAddr = dnode.Addr
			break
		}
	}
	return nil
}

func (c *Coordinator) UpdateFileSize(args *UpdateFileSizeArgs, reply *UpdateFileSizeReply) error {
	log.Println("UpdateFileSizeReq")
	reply.Done = c.mcon.updateFileSize(args.FileId, args.FileSize)
	return nil
}

func (c *Coordinator) CreateFile(args *CreateFileArgs, reply *CreateFileReply) error {
	log.Println("CreateFileReq")
	var dnodes []DataNode
	dnodes = c.mcon.getServers()
	for _, dnode := range dnodes {
		if dnode.IsAlive == true {
			createArgs := CreateFileArgs_m{UserToken: "Accees Token"}
			createReply := CreateFileReply_m{}
			if ok := c.DialDataNode(dnode.Addr, "DataNode.CreateFile_m", &createArgs, &createReply); ok == nil {
				reply.ServerAddr = dnode.Addr
				reply.FileId = c.mcon.insertFile(args.FolderPath, args.FileName, 0)
				break
			} else {
				reply.ServerAddr = ""
			}
		}
	}
	log.Println("Reply:", reply)
	return nil
}

func (c *Coordinator) RenameFile(args *RenameFileArgs, reply *RenameFileReply) error {
	log.Println("RenameFileReq:", args.OldPath, args.OldName, args.NewPath, args.NewName)
	file := c.mcon.getFile(args.OldPath, args.OldName)
	if file.id == "" {
		folder := c.mcon.getFolder(args.OldPath, args.OldName)
		if folder.folderId == "" {
			reply.Status = false
			return nil
		} else {
			// reply.Status = c.mcon.renameFolder(args.OldPath, args.OldName, args.NewPath, args.NewName)
			return nil
		}
	}

	reply.Status = c.mcon.renameFile(args.OldPath, args.OldName, args.NewPath, args.NewName)
	return nil
}

func (c *Coordinator) ListFiles(args *ListFilesArgs, reply *ListFilesReply) error {
	// c.mcon.getFilesFromFolder("/home/test1/Desktop")
	// c.mcon.getFile("/home/test1/Desktop", "test1.txt")
	log.Println("ListFilesReq:", args)
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
	log.Println("ListFilesReply:", reply)
	return nil
}

func (c *Coordinator) InsertFolder(args *InsertFolderArgs, reply *InsertFolderReply) error {
	log.Println("InsertFolderReq")
	reply.FolderId, reply.FolderModified = c.mcon.insertFolder(args.ParentPath, args.FolderName)
	if reply.FolderId == "" {
		reply.FolderId = "0"
	}
	return nil
}

func (c *Coordinator) DeleteFolder(args *DeleteFolderArgs, reply *DeleteFolderReply) error {
	log.Println("DeleteFolderReq")
	reply.DeleteCount = c.mcon.deleteFolder(args.ParentPath, args.FolderName)
	return nil
}
