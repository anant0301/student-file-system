package main

import (
	"context"
	"fmt"
	"math/big"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type Client struct {
	// Must embed an Inode for the struct to work as a node.
	fs.Inode      /* does the client have only one inodes? isn't like inode one for each file */
	file     File /* File struct */
	utime    time.Time
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

func (client *Client) getattr(out *fuse.AttrOut) {
	out.Size = uint64(100)
	fmt.Println("")
	out.SetTimes(nil, &client.utime, nil)
}

// Implement GetAttr to provide size and mtime
var _ = (fs.NodeGetattrer)((*Client)(nil))

func (client *Client) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	// bn.mu.Lock()
	// defer bn.mu.Unlock()
	client.getattr(out)
	return 0
}

// Ensure we are implementing the NodeReaddirer interface
var _ = (fs.NodeReaddirer)((*Client)(nil))

// Readdir is part of the NodeReaddirer interface
func (n *Client) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r := make([]fuse.DirEntry, 0)

	for _, val := range getDir() {
		i := new(big.Int)
		i.SetString(val.FileId, 16)
		fmt.Println("Val:", val)
		if val.IsFolder {
			d := fuse.DirEntry{
				Name: val.FileName,
				Ino:  uint64(val.FileId), // Should be id of the file/ directory
				Mode: fuse.S_IFDIR,       // folder
			}
			r = append(r, d)
		} else {
			d := fuse.DirEntry{
				Name: val.FileName,
				Ino:  uint64(val.FileId), // Should be id of the file/ directory
				Mode: fuse.S_IFREG,       // file
			}
			r = append(r, d)
		}
	}
	return fs.NewListDirStream(r), 0
}
