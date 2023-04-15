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

type Client struct {
	// Must embed an Inode for the struct to work as a node.
	fs.Inode
	mu      sync.Mutex
	file    File
	content []byte
}

const (
	FILE   = false
	FOLDER = true
)

type File struct {
	name     string
	id       uint64
	fileType bool
	sz       uint64
}

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

func getFile(name string) (File, error) {
	for i := range files {
		if files[i].name == name {
			return files[i], nil
		}
	}
	return File{}, errors.New("File not found")
}

// Mode for reference
// fuse.S_IFREG = File
// fuse.S_IFDIR = directories

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReaddirer)((*Client)(nil))

// Readdir is part of the NodeReaddirer interface
func (n *Client) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r := make([]fuse.DirEntry, 0)

	// files := []string{"abc", "bcd", "def", "efg", "fgh"}
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

// Ensure we are implementing the NodeLookuper interface
var _ = (fs.NodeLookuper)((*Client)(nil))

// Lookup is part of the NodeLookuper interface“
func (n *Client) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	// Get the file from the server
	file, err := getFile(n.file.name + "/" + name)

	fmt.Println("File:", file)
	fmt.Println(err)

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

	operations := &Client{file: File{name: file.name, id: file.id, fileType: file.fileType, sz: n.file.sz}}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)

	// In case of concurrent lookup requests, it can happen that operations !=
	// child.Operations().
	return child, 0
}

var _ = (fs.NodeOpener)((*Client)(nil))

func (n *Client) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	// download file from server
	// fd, err := os.Open("/home/hari/Documents/Projects/student-file-system/files/" + n.file.name)
	// syscall.Dup(int(fd.Fd()))
	// if err != nil {
	// 	log.Fatal("Error Opening File")
	// }
	// // fh = fs.NewLoopbackFile(int(fd.Fd()))
	return nil, 0, 0
}

// Implement handleless read.
var _ = (fs.NodeReader)((*Client)(nil))

func (bn *Client) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	// end := off + int64(len(dest))
	// if end > int64(len(bn.content)) {
	// 	end = int64(len(bn.content))
	// }

	// We could copy to the `dest` buffer, but since we have a
	// []byte already, return that.
	copy(dest, bn.content[off:])
	// return , 0
	return fuse.ReadResultData(dest), 0
}

func (bn *Client) resize(sz uint64) {
	if sz > uint64(cap(bn.content)) {
		n := make([]byte, sz)
		copy(n, bn.content)
		bn.content = n
	} else {
		bn.content = bn.content[:sz]
	}
}

// Implement GetAttr to provide size and mtime
var _ = (fs.NodeGetattrer)((*Client)(nil))

func (n *Client) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	// bn.mu.Lock()
	// defer bn.mu.Unlock()
	n.getattr(out)
	return 0
}

func (n *Client) getattr(out *fuse.AttrOut) {
	out.Size = uint64(n.file.sz)
	out.Owner = fuse.Owner{Uid: 1000, Gid: 1000}
	out.SetTimes(nil, &time.Time{}, nil)
}

var _ = (fs.NodeFsyncer)((*Client)(nil))

func (n *Client) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return 0
}

// Implement Setattr to support truncation
var _ = (fs.NodeSetattrer)((*Client)(nil))

func (bn *Client) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
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

// Implement handleless write.
var _ = (fs.NodeWriter)((*Client)(nil))

func (bn *Client) Write(ctx context.Context, fh fs.FileHandle, buf []byte, off int64) (uint32, syscall.Errno) {
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

var _ = (fs.NodeCreater)((*Client)(nil))

func (n *Client) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
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
	operations := &Client{file: file}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)

	files = append(files, file)
	return child, nil, 0, 0
}

var _ = (fs.NodeMkdirer)((*Client)(nil))

func (n *Client) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
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
	operations := &Client{file: file}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)
	files = append(files, file)

	return child, 0
}

func main() {
	// This is where we'll mount the FS
	mntDir := "/home/hari/sfs1"
	os.Mkdir(mntDir, 0777)
	root := &Client{file: File{name: "/root", id: 0, fileType: FOLDER}}
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
