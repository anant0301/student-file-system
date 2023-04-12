package main

import (
	"fmt"
	"log"
	"net/rpc"
)

func getFileTest(client *rpc.Client, filename string) {
	args := GetFileArgs{
		FileName:   filename,
		FolderPath: "/home/test1/Desktop",
	}
	reply := GetFileReply{}
	client.Call("Coordinator.GetFile", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func insertFileTest(client *rpc.Client, filename string) {
	args := InsertFileArgs{
		FileName:   filename,
		FolderPath: "/home/test1/Desktop",
		FileSize:   100,
	}
	reply := InsertFileReply{}
	client.Call("Coordinator.InsertFile", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func listFilesTest(client *rpc.Client, folderPath string) {
	args := ListFilesArgs{
		FolderPath: folderPath,
	}
	reply := ListFilesReply{}
	client.Call("Coordinator.ListFiles", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func deleteFileTest(client *rpc.Client, filename string) {
	args := DeleteFileArgs{
		FileName:   filename,
		FolderPath: "/home/test1/Desktop",
	}
	reply := DeleteFileReply{}
	client.Call("Coordinator.DeleteFile", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func main() {
	client, err := rpc.DialHTTP("tcp", "localhost:9000")
	if err != nil {
		log.Println("Error in dialing rpc:", err)
	}
	insertFileTest(client, "test1.txt")
	insertFileTest(client, "test2.txt")
	getFileTest(client, "test1.txt")
	getFileTest(client, "test2.txt")
	listFilesTest(client, "/home/test1/Desktop")
	deleteFileTest(client, "test1.txt")

}
