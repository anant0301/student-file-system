package main

import (
	"fmt"
	"log"
	"net/rpc"
)

func getFileTest(master *rpc.Coordinator, filename string) {
	args := GetFileArgs_m{
		FileName:   filename,
		FolderPath: "/home/test1/Desktop",
	}
	reply := GetFileReply_m{}
	master.Call("DataNode.GetFile_m", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func insertFileTest(master *rpc.Coordinator, filename string) {
	args := InsertFileArgs_m{
		FileName:   filename,
		FolderPath: "/home/test1/Desktop",
		FileSize:   100,
	}
	reply := InsertFileReply_m{}
	master.Call("DataNode.InsertFile_m", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func deleteFileTest(master *rpc.Coordinator, filename string) {
	args := DeleteFileArgs_m{
		FileName:   filename,
		FolderPath: "/home/test1/Desktop",
	}
	reply := DeleteFileReply_m{}
	master.Call("DataNode.DeleteFile_m", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func (c *Coordinator) InsertFile_done() {
	//update database about insert file at all datanodes specified
}

func main() {
	master, err := rpc.DialHTTP("tcp", "localhost:9000")
	if err != nil {
		log.Println("Error in dialing rpc:", err)
	}
	insertFileTest(master, "test1.txt")
	insertFileTest(master, "test2.txt")
	getFileTest(master, "test1.txt")
	getFileTest(master, "test2.txt")
	deleteFileTest(master, "test1.txt")

}
