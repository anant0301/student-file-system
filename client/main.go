package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Client struct {
	// Must embed an Inode for the struct to work as a node.
	fs.Inode
	id int
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

	for _, val := range getDir(n.id) {
		d := fuse.DirEntry{
			Name: strconv.Itoa(val),
			Ino:  uint64(val), // Should be id of the file/ directory
			Mode: fuse.S_IFDIR,
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
		Mode: fuse.S_IFDIR,
		// The child inode is identified by its Inode number.
		// If multiple concurrent lookups try to find the same
		// inode, they are deduplicated on this key.
		Ino: uint64(i),
	}

	operations := &Client{id: i}

	// The NewInode call wraps the `operations` object into an Inode.
	child := n.NewInode(ctx, operations, stable)

	// In case of concurrent lookup requests, it can happen that operations !=
	// child.Operations().
	return child, 0
}

func main() {
	// This is where we'll mount the FS
	mntDir := "/tmp/y"
	os.Mkdir(mntDir, 0755)
	root := &Client{id: 20}
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
