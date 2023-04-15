package main

// DOCS
// https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs#FileHandle

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
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
	name     string
	id       uint64
	fileType bool
	sz       uint64
}

// Temporary initial files used for testing
var files = []File{
	{name: "/root/abc.txt", id: 1, fileType: FILE},
	{name: "/root/bcd.txt", id: 2, fileType: FILE},
	{name: "/root/def.txt", id: 3, fileType: FILE},
	{name: "/root/efg.txt", id: 4, fileType: FILE},
	{name: "/root/fgh.txt", id: 5, fileType: FILE},
}

// A temporary function that sends list of entries in a directory
func getDir() []File {
	return files
}

// A temporary function that sends the file with the given name
func getFile(name string) (File, error) {
	for i := range files {
		if files[i].name == name {
			return files[i], nil
		}
	}
	return File{}, errors.New("File not found")
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

	for _, val := range getDir() {
		var mode uint32
		if val.fileType == FOLDER {
			mode = fuse.S_IFDIR
		} else {
			mode = fuse.S_IFREG
		}
		slices := strings.Split(val.name, "/")
		name := slices[len(slices)-1]
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
	file, err := getFile(n.file.name + "/" + name)

	if err != nil {
		return nil, syscall.ENOENT
	}

	var mode uint32 = fuse.S_IFREG
	if file.fileType == FOLDER {
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
	operations := &FSNode{file: File{name: file.name, id: file.id, fileType: file.fileType, sz: n.file.sz}}

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
func (bn *FSNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	copy(dest, bn.content[off:])

	return fuse.ReadResultData(dest), 0
}

// To Resize the temp content buffer in FSNode
func (bn *FSNode) resize(sz uint64) {
	if sz > uint64(cap(bn.content)) {
		n := make([]byte, sz)
		copy(n, bn.content)
		bn.content = n
	} else {
		bn.content = bn.content[:sz]
	}
}

// Implement GetAttr to provide size and mtime
var _ = (fs.NodeGetattrer)((*FSNode)(nil))

// Getattr is part of the NodeGetattrer interface.
func (n *FSNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	// bn.mu.Lock()
	// defer bn.mu.Unlock()
	n.getattr(out)
	return 0
}

// getattr fills out the AttrOut structure.
func (n *FSNode) getattr(out *fuse.AttrOut) {
	out.Size = uint64(n.file.sz)
	out.Owner = fuse.Owner{Uid: 1000, Gid: 1000}

	// setting last modified time
	out.SetTimes(nil, &time.Time{}, nil)
}

// Implement Fsync to support file sync
var _ = (fs.NodeFsyncer)((*FSNode)(nil))

// Fsync is part of the NodeFsyncer interface.
// Doing nothing as of now, but can be used to sync the file to the server.
func (n *FSNode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return 0
}

// Implement Setattr to support truncation of file.
var _ = (fs.NodeSetattrer)((*FSNode)(nil))

// Setattr is part of the NodeSetattrer interface.
// It is used to truncate the file.
func (bn *FSNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	fmt.Println("Setattr:", in)
	if sz, ok := in.GetSize(); ok {
		fmt.Println("Size:", sz)
		bn.resize(sz)
		bn.file.sz = uint64(sz)
	}
	bn.getattr(out)
	return 0
}

// To Implement handleless write.
var _ = (fs.NodeWriter)((*FSNode)(nil))

// Write is part of the NodeWriter interface.
// It writes the content of the file to the server. Offset defines the starting point of the write.
func (bn *FSNode) Write(ctx context.Context, fh fs.FileHandle, buf []byte, off int64) (uint32, syscall.Errno) {
	bn.mu.Lock()
	defer bn.mu.Unlock()
	fmt.Println("Offset:", off)
	fmt.Println("Buff", buf)
	sz := int64(len(buf))

	fmt.Println("Size in Write:", sz)
	bn.file.sz = uint64(off + sz)
	bn.resize(uint64(off + sz))
	copy(bn.content[off:off+sz], buf)
	return uint32(sz), 0
}

// To Implement Create
var _ = (fs.NodeCreater)((*FSNode)(nil))

// Create is part of the NodeCreater interface. It is called when a new file is created.
func (n *FSNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	fmt.Println("Create:", name)
	fmt.Println("Flags:", flags)
	fmt.Println("Mode:", mode)
	fmt.Println("Out:", out)

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

	files = append(files, file)
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
	files = append(files, file)

	return child, 0
}

func main() {
	// This is where we'll mount the FS
	mntDir := "/home/hari/sfs1"
	os.Mkdir(mntDir, 0777)
	root := &FSNode{file: File{name: "/root", id: 0, fileType: FOLDER}}
	server, err := fs.Mount(mntDir, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			// Set to true to see how the file system works.
			AllowOther: true,
			Debug:      true,
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
