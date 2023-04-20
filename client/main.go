package main

// REFERENCE DOCS
// https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Constants that define the file type
const (
	FILE   = false
	FOLDER = true
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
}

func getDir(path string) []File {
	files := make([]File, 0)
	listFilesArgs := ListFilesArgs{UserToken: "abc", FolderPath: path}
	listFilesReply := ListFilesReply{}

	err := Call("Coordinator.ListFiles", listFilesArgs, &listFilesReply)

	if err != nil {
		fmt.Println("Coordinator.ListFiles error: ", err)
	}

	fmt.Println("ListFilesReply: ", listFilesReply)

	for _, file := range listFilesReply.Files {
		id, _ := strconv.ParseUint(file.FileId[8:], 16, 64)
		f := File{name: file.FileName, id: id, fileType: file.IsFolder, sz: uint64(file.FileSize), idString: file.FileId}
		files = append(files, f)
	}

	return files
}

func getFilePath(file File) string {
	return file.parentPath + "/" + file.name
}

// A temporary function that sends the file with the given name
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

	for _, val := range getDir(getFilePath(n.file)) {
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

	copy(dest, n.content[off:])

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
	out.Size = uint64(n.file.sz)
	out.Owner = fuse.Owner{Uid: 1000, Gid: 1000}

	// setting last modified time
	// out.SetTimes(nil, &n.file.modifiedTime, nil)

	return 0
}

// Implement Setattr to support truncation of file.
var _ = (fs.NodeSetattrer)((*FSNode)(nil))

// Setattr is part of the NodeSetattrer interface.
// It is used to truncate the file.
func (n *FSNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	n.mu.Lock()
	defer n.mu.Unlock()

	fmt.Println("Setattr:", in)
	if sz, ok := in.GetSize(); ok {
		n.resize(sz)
		n.file.sz = uint64(sz)
	}

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

	sz := int64(len(buf))
	n.file.sz = uint64(off + sz)
	n.resize(uint64(off + sz))
	copy(n.content[off:off+sz], buf)
	return uint32(sz), 0
}

// Implement Fsync to support file sync
var _ = (fs.NodeFsyncer)((*FSNode)(nil))

// Fsync is part of the NodeFsyncer interface.
// Doing nothing as of now, but can be used to sync the file to the server.
func (n *FSNode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return 0
}

// To Implement Create
var _ = (fs.NodeCreater)((*FSNode)(nil))

// Create is part of the NodeCreater interface. It is called when a new file is created.
func (n *FSNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	id := time.Now().Unix()
	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: uint64(id),
	}

	fileType := FILE

	file := File{name: n.file.name + "/" + name, id: uint64(id), fileType: fileType, sz: 0}
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
	id := time.Now().Unix()

	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: uint64(id),
	}

	file := File{name: n.file.name + "/" + name, id: uint64(id), fileType: fileType, sz: 0}
	operations := &FSNode{file: file}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)
	// files = append(files, file)

	return child, 0
}

func main() {
	// This is where we'll mount the FS
	mntDir := "/tmp/sfs"
	os.Mkdir(mntDir, 0777)
	root := &FSNode{file: File{name: "test1", parentPath: "/home", id: 0, fileType: FOLDER}}
	server, err := fs.Mount(mntDir, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			AllowOther: true,
			// Set to true to see how the file system works.
			Debug: true,
		},
	})
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Mounted on %s", mntDir)
	log.Printf("Unmount by calling 'fusermount -u %s'", mntDir)

	// Wait until unmount before exiting
	server.Wait()

	// FOR TESTING RPC CALLS

	// a := new(big.Int)
	// ListFilesArgs := GetFileArgs{UserToken: "abc", FileId: *a, FileName: "abc.txt", FolderPath: "/root"}
	// ListFilesReply := GetFileReply{}

	// fmt.Println(ListFilesReply)

	// err := Call("Coordinator.GetFile", &ListFilesArgs, &ListFilesReply)

	// fmt.Println(err)
	// fmt.Println(ListFilesReply)
}
