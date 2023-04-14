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
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Client struct {
	// Must embed an Inode for the struct to work as a node.
	fs.Inode
	file File
}

const (
	FILE   = false
	FOLDER = true
)

type File struct {
	name     string
	id       int
	fileType bool
}

// A temporary function that sends list of entries in a directory
func getDir() []File {
	files := []File{
		{name: "/root/abc.txt", id: 1, fileType: FILE},
		{name: "/root/bcd.txt", id: 2, fileType: FILE},
		{name: "/root/def.txt", id: 3, fileType: FILE},
		{name: "/root/efg.txt", id: 4, fileType: FILE},
		{name: "/root/fgh.txt", id: 5, fileType: FILE},
	}

	return files
}

func getFile(name string) (File, error) {
	files := []File{
		{name: "/root/abc.txt", id: 1, fileType: FILE},
		{name: "/root/bcd.txt", id: 2, fileType: FILE},
		{name: "/root/def.txt", id: 3, fileType: FILE},
		{name: "/root/efg.txt", id: 4, fileType: FILE},
		{name: "/root/fgh.txt", id: 5, fileType: FILE},
	}

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
		slices := strings.Split(val.name, "/")
		name := slices[len(slices)-1]
		d := fuse.DirEntry{
			Name: name,
			Ino:  uint64(val.id), // Should be id of the file/ directory
			Mode: fuse.S_IFREG,
		}
		r = append(r, d)
	}
	return fs.NewListDirStream(r), 0
}

// Ensure we are implementing the NodeLookuper interface
var _ = (fs.NodeLookuper)((*Client)(nil))

// Lookup is part of the NodeLookuper interface“
func (n *Client) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fmt.Println("Lookup:", n.file.name)
	fmt.Println("Name:", name)
	fmt.Println("Context:", ctx)
	file, err := getFile("/" + n.file.name + "/" + name)

	fmt.Println("File:", file)
	fmt.Println(err)

	if err != nil {
		return nil, syscall.ENOENT
	}

	fmt.Println("File ID:", file.id)
	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: uint64(int(file.id)),
	}

	operations := &Client{file: File{name: file.name, id: file.id, fileType: file.fileType}}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)

	// In case of concurrent lookup requests, it can happen that operations !=
	// child.Operations().
	return child, 0
}

var _ = (fs.NodeOpener)((*Client)(nil))

func (n *Client) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	fd, err := os.Open("/home/anant/Desktop/Anant/student-file-system/files/" + n.file.name)
	syscall.Dup(int(fd.Fd()))
	if err != nil {
		log.Fatal("Error Opening File")
	}
	fh = fs.NewLoopbackFile(int(fd.Fd()))
	return fh, 0, 0
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
	out.Size = uint64(100)
	out.SetTimes(nil, &time.Time{}, nil)
}

func main() {
	// This is where we'll mount the FS
	mntDir := "/tmp/z"
	os.Mkdir(mntDir, 0755)
	root := &Client{file: File{name: "root", id: 0, fileType: FOLDER}}
	server, err := fs.Mount(mntDir, root, &fs.Options{
		MountOptions: fuse.MountOptions{
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
}
