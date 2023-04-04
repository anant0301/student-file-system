package main

// DOCS
// https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs#FileHandle

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Client struct {
	// Must embed an Inode for the struct to work as a node.
	fs.Inode
	id   int
	name string
}

// A temporary function that sends list of entries in a directory
func getDir(id int) []int {
	fmt.Println("From GetDir ", id)
	out := []int{}
	for i := 0; i < id; i++ {
		out = append(out, i)
	}
	return out
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
	for _, val := range getDir(n.id) {
		d := fuse.DirEntry{
			Name: strconv.Itoa(val),
			Ino:  uint64(val), // Should be id of the file/ directory
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
	i, err := strconv.Atoi(name)
	if err != nil {
		return nil, syscall.ENOENT
	}

	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: uint64(i),
	}

	files := []string{"abc", "bcd", "def", "efg", "fgh"}
	operations := &Client{id: i, name: files[i] + ".txt"}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)

	// In case of concurrent lookup requests, it can happen that operations !=
	// child.Operations().
	return child, 0
}

var _ = (fs.NodeOpener)((*Client)(nil))

func (n *Client) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	fd, err := os.Open("/home/hari/Documents/Projects/student-file-system/files/" + n.name)
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
	root := &Client{id: 5}
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
