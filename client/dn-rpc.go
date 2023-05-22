package main
import (
    "fmt"
    "net/rpc"
)

type CreateFileArgs_dn struct {
	FileId string
}

type CreateFileReply_dn struct {
	Status bool
}
type GetFileArgs_c struct {
	AccessToken string
	FileId      string
	Offset      int64
	SizeOfChunk int
}

type GetFileReply_c struct {
	Status bool
	Data   []byte
}


func CallDataNode(addr string, rpcname string, args interface{}, reply interface{}) error {

	c, err := rpc.DialHTTP("tcp", addr)

	if err != nil {
		fmt.Println("Can't connect to server: ", addr, err)
		return err
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)

	return err
}
