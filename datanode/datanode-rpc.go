package main

import (
	"bufio"
	"io/ioutil"
	"net/rpc"
	"os"
)

func (dataNode *DataNode) InsertFile_c(request *InsertFileArgs_c, reply *InsertFileReply_c) error {

	fileWriteHandler, err := os.Create(dataNode.DataDirectory)
	if err != nil {
			//panic(err)
			*reply.Status = false
	}
	defer fileWriteHandler.Close()

	fileWriter := bufio.NewWriter(fileWriteHandler)
	_, err = fileWriter.WriteString(request.Data)
	if err != nil {
			//panic(err)
			*reply.Status = false
	}
	fileWriter.Flush()
	*reply.Status = true

	//return dataNode.forwardForReplication(request, reply)
	for idx, addr := range len(request.ReplicationNodes) {
		if dataNode.forwardForReplication(addr, request, reply) != nil {
			*reply.Status = false
		}
	}

	if *reply.Status != false{
		dataNodeInstance, rpcErr := rpc.Dial("tcp", addr.Host+":"+addr.ServicePort)
		if rpcErr != nil {
			//panic(rpcErr)
			*reply.Status = false
		}
		defer dataNodeInstance.Close()

		rpcErr = dataNodeInstance.Call("Coordinator.InsertFile_done", &request, &reply)
		if rpcErr != nil {
			//panic(rpcErr)
			*reply.Status = false
		}
	}

	return nil
}

func (dataNode *DataNode) forwardForReplication(string addr, request *InsertFileArgs_c, reply *InsertFileReply_c) error {

	dataNodeInstance, rpcErr := rpc.Dial("tcp", addr.Host+":"+addr.ServicePort)
	if rpcErr != nil {
			//panic(rpcErr)
			*reply.Status = false
		}
	defer dataNodeInstance.Close()

	rpcErr = dataNodeInstance.Call("DataNode.InsertFile_c", &request, &reply)
	if rpcErr != nil {
			//panic(rpcErr)
			*reply.Status = false
		}
	return nil
}

func (dataNode *DataNode) InsertFile_m(request *InsertFileArgs_m, reply *InsertFileReply_m) error {
	if dataNode.ClientToken == InsertFileArgs_m.AccessToken {
		*reply.Status = true
	} else {
		*reply.Status = false
	}

	return nil
}

func (dataNode *DataNode) GetFile_c(request *GetFileArgs_c, reply *GetFileReply_c) error {
	dataBytes, err := ioutil.ReadFile(dataNode.DataDirectory + request.BlockId)
	//Check(err)
	*reply = DataNodeData{Data: string(dataBytes)}
	return nil
}

func (dataNode *DataNode) GetFile_m(request *GetFileArgs_m, reply *GetFileReply_m) error {
	//dataBytes, err := ioutil.ReadFile(dataNode.DataDirectory + request.BlockId)
	//Check(err)
	if dataNode.ClientToken == request.AccessToken {
		*reply.Status = true
	} else {
		*reply.Status = false
	}

	return nil
}

func (dataNode *DataNode) DeleteFile_m(request *InsertFileArgs_m, reply *InsertFileReply_m) error {
	if dataNode.ClientToken == InsertFileArgs_m.AccessToken {
		*reply.Status = true
	} else {
		*reply.Status = false
	}

	return nil
}

func (dataNode *DataNode) DeleteFile_c(request *InsertFileArgs_c, reply *InsertFileReply_c) error {

	//fileWriteHandler, err := os.Create(dataNode.DataDirectory)
	if err != nil {
			//panic(err)
			*reply.Status = false
	}
	//defer fileWriteHandler.Close()

	//fileWriter := bufio.NewWriter(fileWriteHandler)
	//_, err = fileWriter.WriteString(request.Data)
	if err != nil {
			//panic(err)
			*reply.Status = false
	}
	//fileWriter.Flush()
	*reply.Status = true

	//return dataNode.forwardForReplication(request, reply)
	for idx, addr := range len(request.ReplicationNodes) {
		if dataNode.forwardForReplication(addr, request, reply) != nil {
			*reply.Status = false
		}
	}

	if *reply.Status != false{
		dataNodeInstance, rpcErr := rpc.Dial("tcp", addr.Host+":"+addr.ServicePort)
		if rpcErr != nil {
			//panic(rpcErr)
			*reply.Status = false
		}
		defer dataNodeInstance.Close()

		rpcErr = dataNodeInstance.Call("Coordinator.DeleteFile_done", &request, &reply)
		if rpcErr != nil {
			//panic(rpcErr)
			*reply.Status = false
		}
		return nil

	}

	return nil
}