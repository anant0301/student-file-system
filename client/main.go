package main

// REFERENCE DOCS
// https://pkg.go.dev/github.com/hanwen/go-fuse/v2/fs

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Constants that define the file type
const (
	FILE   = false
	FOLDER = true
)

var rootPath string

var MasterAddr string

func main() {
	// This is where we'll mount the FS
	if len(os.Args) != 5 {
		fmt.Println("Usage: ./client <master-addr> <mount-dir> <root-parent> <root-name>")
		return
	}

	MasterAddr = os.Args[1]
	mntDir := os.Args[2]
	rootParent := os.Args[3]
	rootName := os.Args[4]

	err := os.Mkdir(mntDir, 0777)
	if err != nil {
		log.Fatal("Error in creating mount directory: ", err)
		return
	}

	root := &FSNode{file: File{name: rootName, parentPath: rootParent, id: 0, fileType: FOLDER}}

	rootPath = rootParent + "/" + rootName
	timeoutTime := time.Duration(1 * time.Second)
	server, err := fs.Mount(mntDir, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			// AllowOther: true,
			// Set to true to see how the file system works.
			Debug:         true,
			DisableXAttrs: true,
			EnableLocks:   true,
		},
		// AttrTimeout:  &timeoutTime,
		EntryTimeout: &timeoutTime,
	})
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Mounted on %s", mntDir)
	log.Printf("Unmount by calling 'fusermount -u %s'", mntDir)

	// Wait until unmount before exiting
	server.Wait()

}
