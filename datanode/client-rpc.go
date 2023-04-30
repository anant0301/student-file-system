// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/rpc"
// )

// func getFileTest(client *rpc.Client, filename string) {
// 	args := GetFileArgs{
// 		FileName:   filename,
// 		FolderPath: "/home/test1/Desktop",
// 	}
// 	reply := GetFileReply{}
// 	client.Call("DataNode.GetFile_c", &args, &reply)
// 	fmt.Println("Args:", args, "reply:", reply)
// }

// func insertFileTest(client *rpc.Client, filename string) {
// 	args := InsertFileArgs{
// 		FileName:   filename,
// 		FolderPath: "/home/test1/Desktop",
// 		FileSize:   100,
// 	}
// 	reply := InsertFileReply{}
// 	client.Call("DataNode.InsertFile_c", &args, &reply)
// 	fmt.Println("Args:", args, "reply:", reply)
// }

// func deleteFileTest(client *rpc.Client, filename string) {
// 	args := DeleteFileArgs{
// 		FileName:   filename,
// 		FolderPath: "/home/test1/Desktop",
// 	}
// 	reply := DeleteFileReply{}
// 	client.Call("DataNode.DeleteFile", &args, &reply)
// 	fmt.Println("Args:", args, "reply:", reply)
// }

// func main() {
// 	client, err := rpc.DialHTTP("tcp", "localhost:9000")
// 	if err != nil {
// 		log.Println("Error in dialing rpc:", err)
// 	}
// 	insertFileTest(client, "test1.txt")
// 	insertFileTest(client, "test2.txt")
// 	getFileTest(client, "test1.txt")
// 	getFileTest(client, "test2.txt")
// 	deleteFileTest(client, "test1.txt")

// }