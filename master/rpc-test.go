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
	}
	reply := InsertFileReply{}
	client.Call("Coordinator.CreateFile", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func listFilesTest(client *rpc.Client, folderPath string) {
	args := ListFilesArgs{
		FolderPath: folderPath,
	}
	reply := ListFilesReply{}
	err := client.Call("Coordinator.ListFiles", &args, &reply)
	if err != nil {
		log.Println("Error in calling ListFiles:", err)
	}
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

func insertFolder(client *rpc.Client, folderPath string, folderName string) {
	args := InsertFolderArgs{
		ParentPath: folderPath,
		FolderName: folderName,
	}
	reply := InsertFolderReply{}
	client.Call("Coordinator.InsertFolder", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func deleteFolder(client *rpc.Client, folderPath string, folderName string) {
	args := DeleteFolderArgs{
		ParentPath: folderPath,
		FolderName: folderName,
	}
	reply := DeleteFolderReply{}
	client.Call("Coordinator.DeleteFolder", &args, &reply)
	fmt.Println("Args:", args, "reply:", reply)
}

func main() {
	client, err := rpc.DialHTTP("tcp", "0.0.0.0:9000")
	if err != nil {
		log.Println("Error in dialing rpc:", err)
	}
	insertFileTest(client, "test1.txt")
	insertFileTest(client, "test2.txt")
	getFileTest(client, "test1.txt")
	getFileTest(client, "test2.txt")
	listFilesTest(client, "/home/test1/Desktop")
	deleteFileTest(client, "test1.txt")
	listFilesTest(client, "/home/test1/Desktop")

	insertFolder(client, "/", "home")
	insertFolder(client, "/home", "test1")
	insertFolder(client, "/home/test1", "Desktop")
	insertFolder(client, "/home/test1/Desktop", "testFolder")
	deleteFolder(client, "/home/test1/Desktop", "testFolder")

}
