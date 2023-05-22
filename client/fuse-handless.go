package main

// REFERENCE DOCS
// https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// File System Node Structure
type FSNode struct {
	// Must embed an Inode for the struct to work as a node.
	fs.Inode
	mu      sync.Mutex
	file    File
	content []byte // for testing - remove later
}

type File struct {
	name         string
	id           uint64
	idString     string
	fileType     bool
	sz           uint64
	parentPath   string
	parentId     string
	modifiedTime time.Time
	dataNodeAddr string
}

func getDir(path string) ([]File, error) {
	files := make([]File, 0)
	listFilesArgs := ListFilesArgs{UserToken: "abc", FolderPath: path}
	listFilesReply := ListFilesReply{}

	err := Call("Coordinator.ListFiles", listFilesArgs, &listFilesReply)

	log.Println("ListFilesReply: ", listFilesReply)
	if err != nil {
		fmt.Println("Coordinator.ListFiles error: ", err)
		return files, err
	}

	fmt.Println("ListFilesReply: ", listFilesReply)

	for _, file := range listFilesReply.Files {
		id, _ := strconv.ParseUint(file.FileId[8:], 16, 64)
		f := File{
			name:         file.FileName,
			id:           id,
			idString:     file.FileId,
			fileType:     file.IsFolder,
			sz:           uint64(file.FileSize),
			modifiedTime: file.FileModified,
		}
		files = append(files, f)
	}

	return files, nil
}

func getFilePath(file File) string {
	return file.parentPath + "/" + file.name
}

func getFileIdFromString(idString string) uint64 {
	id, _ := strconv.ParseUint(idString[8:], 16, 64)
	return id
}

func getFile(parent File, name string) (File, error) {
	fmt.Println("Parent", parent)
	fmt.Println("Name", name)
	getFileArgs := GetFileArgs{UserToken: "abc", FolderPath: getFilePath(parent), FileName: name}
	getFileReply := GetFileReply{File: FileStruct{FileId: "000000000000000000000000"}}
	err := Call("Coordinator.GetFile", getFileArgs, &getFileReply)
	if err != nil {
		// log.Fatal("Coordinator.GetFile error: ", err)
		fmt.Println("Coordinator.GetFile error: ", err)
	}

	if getFileReply.File.FileId == "0" || getFileReply.File.FileId == "000000000000000000000000" {
		return File{}, errors.New("File not found")
	}

	id, _ := strconv.ParseUint(getFileReply.File.FileId[8:24], 16, 64)
	file := File{
		name:         getFileReply.File.FileName,
		id:           id,
		fileType:     getFileReply.File.IsFolder,
		sz:           uint64(getFileReply.File.FileSize),
		idString:     getFileReply.File.FileId,
		parentPath:   getFilePath(parent),
		parentId:     parent.idString,
		modifiedTime: getFileReply.File.FileModified,
		dataNodeAddr: getFileReply.NodeAddr,
	}

	return file, nil
}

// FUSE File Mode for reference
// fuse.S_IFREG = File
// fuse.S_IFDIR = directories

// For ReadDir
var _ = (fs.NodeReaddirer)((*FSNode)(nil))

// Readdir is part of the NodeReaddirer interface
// It returns a list of files in the directory
func (n *FSNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r := make([]fuse.DirEntry, 0)

	dirList, err := getDir(getFilePath(n.file))

	if err != nil {
		fmt.Println("Error in Readdir: ", err)
		return fs.NewListDirStream(r), 0
	}

	for _, val := range dirList {
		var mode uint32

		fmt.Println("File Val: ", val)
		if val.fileType == FOLDER {
			fmt.Println("Folder")
			mode = fuse.S_IFDIR
		} else {
			mode = fuse.S_IFREG
		}
		name := val.name
		d := fuse.DirEntry{
			Name: name,
			Ino:  uint64(val.id), // Should be id of the file/ directory
			Mode: mode,
		}
		r = append(r, d)
	}
	return fs.NewListDirStream(r), 0
}

// For Lookup
var _ = (fs.NodeLookuper)((*FSNode)(nil))

// Lookup is part of the NodeLookuper interface
// It returns the file with the given name
func (n *FSNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	// Get the file from the server
	log.Println("Lookup called !!")
	log.Println("Lookup: ", name, "n:", n)
	file, err := getFile(n.file, name)

	if err != nil {
		return nil, syscall.ENOENT
	}

	var mode uint32 = fuse.S_IFREG
	if file.fileType == FOLDER {
		fmt.Println("Folder Lookup")
		mode = fuse.S_IFDIR
	}

	stable := fs.StableAttr{
		Mode: mode,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: uint64(int(file.id)),
	}

	// Create a new FSNode for the file
	operations := &FSNode{file: file}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)

	// In case of concurrent lookup requests, it can happen that operations !=
	// child.Operations().
	return child, 0
}

var _ = (fs.NodeOpener)((*FSNode)(nil))

func (n *FSNode) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	// The commented code returns the file descriptor of the file

	// fd, err := os.Open("/home/hari/Documents/Projects/student-file-system/files/" + n.file.name)
	// syscall.Dup(int(fd.Fd()))
	// if err != nil {
	// 	log.Fatal("Error Opening File")
	// }
	// // fh = fs.NewLoopbackFile(int(fd.Fd()))

	// Return nil as file handle to indicate that we don't need one.
	return nil, 0, 0
}

// Implement handleless read.
// For Read
var _ = (fs.NodeReader)((*FSNode)(nil))

// Read is part of the NodeReader interface.
// It returns the content of the file as a byte array to the client.
// Offset defines the starting point of the read.
func (n *FSNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	n.mu.Lock()
	defer n.mu.Unlock()

	getFileArgs := GetFileArgs{
		UserToken:  "abc",
		FolderPath: n.file.parentPath,
		FileName:   n.file.name,
	}

	getFileReply := GetFileReply{}

	log.Println("Read Get File Args:", getFileArgs)

	err := Call("Coordinator.GetFile", getFileArgs, &getFileReply)
	if err != nil {
		log.Println("Error in GetFile RPC in Read", err)
		return fuse.ReadResultData(dest), syscall.ECONNABORTED
	}

	getFileArg_c := GetFileArgs_c{
		AccessToken: "abc",
		FileId:      n.file.idString,
		Offset:      int64(off),
		SizeOfChunk: int(len(dest)),
	}
	getFileReply_c := GetFileReply_c{}

	err1 := CallDataNode(getFileReply.NodeAddr, "DataNode.GetFile_c", getFileArg_c, &getFileReply_c)

	if err1 != nil {
		log.Println("Read Get File Reply:", getFileReply)
		log.Println("Error in GetFile_c RPC in Read", err)
		return fuse.ReadResultData(dest), syscall.ECONNABORTED
	}

	copy(dest, getFileReply_c.Data)

	return fuse.ReadResultData(dest), 0
}

// To Resize the temp content buffer in FSNode
func (n *FSNode) resize(sz uint64) {
	if sz > uint64(cap(n.content)) {
		new := make([]byte, sz)
		copy(new, n.content)
		n.content = new
	} else {
		n.content = n.content[:sz]
	}
}

// Implement GetAttr to provide size and mtime
var _ = (fs.NodeGetattrer)((*FSNode)(nil))

// Getattr is part of the NodeGetattrer interface.
func (n *FSNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	n.mu.Lock()
	defer n.mu.Unlock()
	log.Println("Getattr:", n.file.name)
	out.Size = uint64(n.file.sz)
	out.Owner = fuse.Owner{Uid: 1000, Gid: 1000}

	// setting last modified time
	out.SetTimes(nil, &n.file.modifiedTime, nil)

	return 0
}

// Implement Setattr to support truncation of file.
var _ = (fs.NodeSetattrer)((*FSNode)(nil))

// Setattr is part of the NodeSetattrer interface.
// It is used to truncate the file.
func (n *FSNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	n.mu.Lock()
	defer n.mu.Unlock()

	// fmt.Println("Setattr:", in)
	// if sz, ok := in.GetSize(); ok {
	// 	n.resize(sz)
	// 	n.file.sz = uint64(sz)
	// }

	out.Size = uint64(n.file.sz)
	out.Owner = fuse.Owner{Uid: 1000, Gid: 1000}

	// setting last modified time
	out.SetTimes(nil, &time.Time{}, nil)

	return 0
}

// To Implement handleless write.
var _ = (fs.NodeWriter)((*FSNode)(nil))

// Write is part of the NodeWriter interface.
// It writes the content of the file to the server. Offset defines the starting point of the write.
func (n *FSNode) Write(ctx context.Context, fh fs.FileHandle, buf []byte, off int64) (uint32, syscall.Errno) {
	n.mu.Lock()
	defer n.mu.Unlock()

	getFileArgs := GetFileArgs{
		UserToken:  "abc",
		FolderPath: n.file.parentPath,
		FileName:   n.file.name,
	}

	getFileReply := GetFileReply{}
	log.Println("Write Get File Args:", getFileArgs)
	log.Println("Write Get File Offset:", off)

	err := Call("Coordinator.GetFile", getFileArgs, &getFileReply)
	if err != nil {
		log.Println("Error in GetFile RPC in Write", err)
		return 0, syscall.ECONNABORTED
	}

	insertFileArgs_c := InsertFileArgs_c{
		Data:   buf,
		Offset: int64(off),
		FileId: n.file.idString,
	}
	insertFileReply_c := InsertFileReply_c{}

	log.Println("Write Insert File Args:", len(insertFileArgs_c.Data))

	log.Println("GetFile Reply: ", getFileReply)
	err = CallDataNode(getFileReply.NodeAddr, "DataNode.InsertFile_c", insertFileArgs_c, &insertFileReply_c)

	if err != nil {
		log.Println("Error in InsertFile RPC in Write 123", err)
		return 0, syscall.ECONNABORTED
	}

	if insertFileReply_c.Status == false {
		log.Println("Error in InsertFile RPC in Write: server replied false")
		return 0, syscall.ECONNABORTED
	}

	log.Println("Write Success:", off)
	log.Println("Write: ", n.file.sz)

	n.file.sz = uint64(insertFileReply_c.FileSize)

	log.Println("Write Reply: ", insertFileReply_c)

	log.Println("Write: File Size ", n.file.sz)
	return uint32(n.file.sz), 0
}

// Implement Fsync to support file sync
var _ = (fs.NodeFsyncer)((*FSNode)(nil))

// Fsync is part of the NodeFsyncer interface.
// Doing nothing as of now, but can be used to sync the file to the server.
func (n *FSNode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return 0
}

// var _ = (fs.NodeGetxattrer)((*FSNode)(nil))

// func (n *FSNode) Getxattr(ctx context.Context, attr string, dest []byte) (uint32, syscall.Errno) {
// 	log.Println("Getxattr:", attr)
// 	log.Println("Getxattr: Dest", dest)
// 	return 0, 0
// }

// To Implement Create
var _ = (fs.NodeCreater)((*FSNode)(nil))

// Create is part of the NodeCreater interface. It is called when a new file is created.
func (n *FSNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	// Create a new file in the datanode
	createFileArgs_m := CreateFileArgs_m{UserToken: "abc", FolderPath: getFilePath(n.file), FileName: name}
	createFileReply_m := CreateFileReply_m{}

	err_m := Call("Coordinator.CreateFile", createFileArgs_m, &createFileReply_m)

	if err_m != nil {
		fmt.Println("Error in creating file in master", err_m)
	}

	fmt.Println("File created in master", createFileReply_m)
	idString := createFileReply_m.FileId

	createFileArgs_dn := CreateFileArgs_dn{FileId: idString}
	createFileReply_dn := CreateFileReply_dn{}

	err := CallDataNode(createFileReply_m.ServerAddr, "DataNode.CreateFile_c", createFileArgs_dn, &createFileReply_dn)

	if err != nil {
		fmt.Println("Error in creating file in datanode", err)
	}
	fmt.Println("File created in dn", createFileReply_dn)

	id := getFileIdFromString(idString)
	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: id,
	}

	fileType := FILE

	file := File{
		name:         name,
		id:           id,
		idString:     idString,
		fileType:     fileType,
		sz:           0,
		parentPath:   getFilePath(n.file),
		parentId:     n.file.idString,
		modifiedTime: time.Now(),
	}
	operations := &FSNode{file: file}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)

	// files = append(files, file)
	return child, nil, 0, 0
}

// To Implement Mkdir
var _ = (fs.NodeMkdirer)((*FSNode)(nil))

// Mkdir is part of the NodeMkdirer interface. It is called when a new folder is created.
func (n *FSNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fileType := FOLDER
	insertFolderArgs := InsertFolderArgs{FolderName: name, ParentPath: getFilePath(n.file), UserToken: "123"}
	insertFolderReply := InsertFolderReply{}
	err := Call("Coordinator.InsertFolder", insertFolderArgs, &insertFolderReply)
	if err != nil {
		fmt.Println("Mkdir: Error in calling Coordinator.InsertFolder", err)
	}

	id := getFileIdFromString(insertFolderReply.FolderId)

	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: id,
	}

	file := File{
		name:         name,
		id:           id,
		idString:     insertFolderReply.FolderId,
		fileType:     fileType,
		sz:           0,
		parentPath:   getFilePath(n.file),
		parentId:     n.file.idString,
		modifiedTime: insertFolderReply.FolderModified,
	}

	operations := &FSNode{file: file}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)
	// files = append(files, file)

	return child, 0
}

var _ = (fs.NodeUnlinker)((*FSNode)(nil))

func (n *FSNode) Unlink(ctx context.Context, name string) syscall.Errno {
	fmt.Println("Unlink:", name)
	fmt.Println("Unlink: File:", n.file)

	deleteFileArgs := DeleteFileArgs{
		UserToken:  "123",
		FolderPath: getFilePath(n.file),
		FileName:   name,
	}

	fmt.Println("Unlink: DeleteFileArgs:", deleteFileArgs)
	deleteFileReply := DeleteFileReply{}

	err := Call("Coordinator.DeleteFile", deleteFileArgs, &deleteFileReply)

	if err != nil {
		fmt.Println("Unlink: Error in calling Coordinator.DeleteFile", err)
	}

	log.Println("Unlink: DeleteFileReply:", deleteFileReply)

	return 0
}

var _ = (fs.NodeRmdirer)((*FSNode)(nil))

func (n *FSNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	fmt.Println("Rmdir:", name)
	fmt.Println("Rmdir: File:", n.file)
	deleteFolderArgs := DeleteFolderArgs{FolderName: name, ParentPath: getFilePath(n.file), UserToken: "123"}
	deleteFolderReply := DeleteFolderReply{}

	err := Call("Coordinator.DeleteFolder", deleteFolderArgs, &deleteFolderReply)
	if err != nil {
		fmt.Println("Rmdir: Error in calling Coordinator.DeleteFolder", err)
	}

	return 0
}

var _ = (fs.NodeRenamer)((*FSNode)(nil))

func (n *FSNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	log.Println("Rename: Name", name)
	log.Println("Rename: NewName", newName)

	entireParentName := ""
	parentPointer := newParent.EmbeddedInode()

	for parentPointer != nil {
		parentName, p := parentPointer.Parent()
		parentPointer = p
		entireParentName = parentName + "/" + entireParentName
	}
	entireParentName = rootPath + entireParentName
	log.Println("Rename: ParentName", entireParentName)
	newParentName := entireParentName[:len(entireParentName)-1]
	oldParentName := getFilePath(n.file)

	args := RenameFileArgs{
		OldPath: oldParentName,
		NewPath: newParentName,
		OldName: name,
		NewName: newName,
	}

	reply := RenameFileReply{}
	log.Println("Rename: Args", args)
	err := Call("Coordinator.RenameFile", args, &reply)

	if err != nil {
		log.Println("Rename: Error in calling Coordinator.RenameFile", err)
	}

	log.Println("Rename: Reply", reply)
	// newParent.EmbeddedInode().NotifyEntry(newName)

	return 0
}
