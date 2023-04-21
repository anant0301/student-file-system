package main

import (
	"fmt"
	"net/rpc"
	"os"
)

var MasterAddr = "10.7.50.133:9000"
var Me = "10.0.60.100:9000"

func (dataNode *DataNode) Ping() error {
	dataNodeInstance, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
	if rpcErr != nil {
		//panic(rpcErr)
		// reply.Success = false
		return rpcErr
	}
	defer dataNodeInstance.Close()

	request := PingArgs{}
	reply := PingReply{}
	request.Addr = "10.0.60.100:9000"

	fmt.Println("Ping to master")
	// request2:=*GetReplicationNodes_Args
	// reply2:=*GetReplicationNodes_Reply

	rpcErr = dataNodeInstance.Call("Coordinator.Ping", request, &reply)
	if rpcErr != nil {
		//panic(rpcErr)
		reply.Status = false
		//return rpcErr
	}

	//fmt.Println(reply.Status)
	return nil
}

func (dataNode *DataNode) CreateFile_c(request *CreateFileArgs_c, reply *CreateFileReply_c) error {
	fmt.Println("Got Create file request from client")
	// err := os.Remove(request.FileId)
	// if err != nil {
	// 	//panic(err)
	// 	reply.Status = false
	// 	return err
	// }
	name, err2 := os.Create(request.FileId)
	fmt.Println(err2, name)
	if err2 != nil {
		//panic(err)
		reply.Status = false
		return err2
	}
	fmt.Println("File created")
	err := dataNode.forwardForReplicationCreate(request, reply)
	if err != nil {
		reply.Status = false
		//return err
	}
	//reply.Status = true
	fmt.Println(reply.Status)
	return nil
}

func (dataNode *DataNode) InsertFile_c(request *InsertFileArgs_c, reply *InsertFileReply_c) error {
	NoOfBytes := 0
	if request.Offset == 0 {
		err := os.Remove(request.FileId)
		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		file1, err2 := os.Create(request.FileId)
		if err2 != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		NoOfBytes, err = file1.WriteAt(request.Data, request.Offset)

		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		reply.Status = true

	} else {
		file, err := os.Open(dataNode.DataDirectory + request.FileId)
		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		defer file.Close()

		//fileWriter := bufio.NewWriter(fileWriteHandler)

		NoOfBytes, err = file.WriteAt(request.Data, request.Offset)

		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		reply.Status = true
	}

	//return dataNode.forwardForReplication(request, reply)

	err := dataNode.forwardForReplicationInsert(int64(NoOfBytes), request, reply)
	if err != nil {
		reply.Status = false
		//return err
	}

	// if reply.Status != false {
	// 	dataNodeInstance, rpcErr := rpc.Dial("tcp", MasterAddr.Host+":"+MasterAddr.ServicePort)
	// 	if rpcErr != nil {
	// 		//panic(rpcErr)
	// 		reply.Status = false
	// 		return rpcErr
	// 	}
	// 	defer dataNodeInstance.Close()
	// 	request.BytesWritten = NoOfBytes
	// 	rpcErr = dataNodeInstance.Call("Coordinator.InsertFile_done", &request, &reply)
	// 	if rpcErr != nil {
	// 		//panic(rpcErr)
	// 		reply.Status = false
	// 		return rpcErr
	// 	}

	// }

	return nil
}

func (dataNode *DataNode) forwardedForReplicationInsert(request *InsertFileArgs_c, reply *InsertFileReply_c) error {

	if request.Offset == 0 {
		err := os.Remove(request.FileId)
		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		file1, err2 := os.Create(request.FileId)
		if err2 != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		_, err = file1.WriteAt(request.Data, request.Offset)

		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		reply.Status = true

	} else {
		file, err := os.Open(dataNode.DataDirectory + request.FileId)
		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		defer file.Close()

		//fileWriter := bufio.NewWriter(fileWriteHandler)

		_, err = file.WriteAt(request.Data, request.Offset)

		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		reply.Status = true
	}
	return nil

}

func (dataNode *DataNode) forwardForReplicationInsert(NoOfBytes int64, request *InsertFileArgs_c, reply *InsertFileReply_c) error {

	dataNodeInstance2, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
	if rpcErr != nil {
		//panic(rpcErr)
		reply.Status = false
		return rpcErr
	}
	defer dataNodeInstance2.Close()

	request2 := GetReplicationNodesArgs{}
	reply2 := GetReplicationNodesReply{}
	// request2:=*GetReplicationNodes_Args
	// reply2:=*GetReplicationNodes_Reply

	rpcErr = dataNodeInstance2.Call("Coordinator.GetReplicationNodes", &request2, &reply2)
	if rpcErr != nil {
		//panic(rpcErr)
		reply.Status = false
		return rpcErr
	}
	reply.Status = true
	for _, addr := range reply2.ReplicationNodes {
		if addr == Me {
			continue
		}
		dataNodeInstance, rpcErr := rpc.DialHTTP("tcp", addr)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}
		defer dataNodeInstance.Close()

		rpcErr = dataNodeInstance.Call("DataNode.forwardedForReplicationInsert", &request, &reply)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			break
			//return rpcErr
		}

	}
	if !reply.Status {
		dataNodeInstance3, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}
		defer dataNodeInstance3.Close()

		request3 := InsertFileDoneArgs{}
		request3.BytesWritten = NoOfBytes
		reply3 := InsertFileDoneReply{}

		rpcErr = dataNodeInstance3.Call("Coordinator.InsertFileDone", &request3, &reply3)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}

	}

	return nil
}

func (dataNode *DataNode) forwardedForReplicationCreate(request *InsertFileArgs_c, reply *InsertFileReply_c) error {

	err := os.Remove(request.FileId)
	if err != nil {
		//panic(err)
		reply.Status = false
		return err
	}
	_, err2 := os.Create(request.FileId)
	if err2 != nil {
		//panic(err)
		reply.Status = false
		return err
	}

	return nil

}

func (dataNode *DataNode) forwardForReplicationCreate(request *CreateFileArgs_c, reply *CreateFileReply_c) error {
	fmt.Println("forwardForReplication")
	dataNodeInstance2, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
	if rpcErr != nil {
		//panic(rpcErr)
		fmt.Println("sent request replication Nodes", rpcErr)
		reply.Status = false
		return nil
	}
	defer dataNodeInstance2.Close()

	request2 := GetReplicationNodesArgs{}
	reply2 := GetReplicationNodesReply{}
	// request2:=*GetReplicationNodes_Args
	// reply2:=*GetReplicationNodes_Reply

	rpcErr3 := dataNodeInstance2.Call("Coordinator.GetReplicationNodes", &request2, &reply2)
	if rpcErr3 != nil {
		//panic(rpcErr)
		//reply2.Status = false
		//return rpcErr
		reply.Status=false
		return nil
	}
	fmt.Println("got replication Nodes")
	fmt.Println(len(reply2.ReplicationNodes))
	reply.Status = true
	for _, addr := range reply2.ReplicationNodes {

		if addr == Me {
			continue
		}
		dataNodeInstance, rpcErr := rpc.DialHTTP("tcp", addr)

		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}
		defer dataNodeInstance.Close()

		rpcErr = dataNodeInstance.Call("DataNode.forwardedForReplicationCreate", &request, &reply)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			break
			//return rpcErr
		}
	}
	fmt.Println(reply.Status)
	if reply.Status ==true {
		fmt.Println("Sending Done to Master")
		dataNodeInstance3, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}
		defer dataNodeInstance3.Close()

		request3 := DoneArgs{}
		request3.FileId = request.FileId
		request3.FileSize = 0
		request3.Operation = "create"
		// request3.BytesWritten = NoOfBytes
		reply3 := DoneReply{}

		rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}
		fmt.Println("Sending Done to Master Successful",reply3.Status)

	}

	return nil
}

func (dataNode *DataNode) GetFile_c(request *GetFileArgs_c, reply *GetFileReply_c) error {
	// dataBytes, err := ioutil.ReadFile(dataNode.DataDirectory + request.BlockId)
	// //Check(err)
	// *reply = DataNodeData{Data: string(dataBytes)}
	// return nil

	file, err := os.Open(dataNode.DataDirectory + request.FileId)
	if err != nil {
		//panic(err)
		reply.Status = false
		return nil
	}
	defer file.Close()

	//fileWriter := bufio.NewWriter(fileWriteHandler)

	_, _ = file.Seek(request.Offset, 0)
	b1 := make([]byte, request.SizeOfChunk)
	n1, err := file.Read(b1)
	if err != nil {
		//panic(err)
		reply.Status = false
	}

	reply.Data = b1[:n1]
	reply.Status = true

	return nil

}

func (dataNode *DataNode) CreateFile_m(request *CreateFileArgs_m, reply *CreateFileReply_m) error {
	fmt.Println("Got Create file request from master")
	if dataNode.ClientToken == request.UserToken {
		reply.Status = true
	} else {
		reply.Status = false
	}

	return nil
}

func (dataNode *DataNode) InsertFile_m(request *InsertFileArgs_m, reply *InsertFileReply_m) error {
	dataNode.ClientToken = request.UserToken
	reply.Status = true

	return nil
}

func (dataNode *DataNode) GetFile_m(request *GetFileArgs_m, reply *GetFileReply_m) error {
	//dataBytes, err := ioutil.ReadFile(dataNode.DataDirectory + request.BlockId)
	//Check(err)
	if dataNode.ClientToken == request.AccessToken {
		reply.Status = true
	} else {
		reply.Status = false
	}

	return nil
}
