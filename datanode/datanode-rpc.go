package main

import (
	"fmt"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

var MasterAddr = "10.7.50.133:9000"

// func (dataNode *DataNode) DialRPCMaster

func (dataNode *DataNode) Ping(host string, port int) error {
	dataNodeInstance, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
	if rpcErr != nil {
		//panic(rpcErr)
		// reply.Success = false
		return rpcErr
	}
	defer dataNodeInstance.Close()

	request := PingArgs{}
	reply := PingReply{}
	//request.Addr = "10.0.60.100:9000"
	request.Addr = host + ":" + strconv.Itoa(port)

	fmt.Println("Ping to master")
	// request2:=*GetReplicationNodes_Args
	// reply2:=*GetReplicationNodes_Reply

	rpcErr = dataNodeInstance.Call("Coordinator.Ping", request, &reply)
	if rpcErr != nil {
		//panic(rpcErr)
		//reply.Status = false
		fmt.Println(rpcErr)
	}
	fmt.Println(reply.Logs)
	if len(reply.Logs) > 0 {
		dataNode.UpdateMyself(&reply)
	}

	//fmt.Println(reply.Status)
	return nil
}

func (dataNode *DataNode) UpdateMyself(reply *PingReply) error {
	fmt.Println("Update myself")
	for _, filelog := range reply.Logs {

		if filelog.Operation == "insert" {

			dataNodeInstance, rpcErr := rpc.DialHTTP("tcp", filelog.Addr)
			if rpcErr != nil {
				//panic(rpcErr)
				return rpcErr
			}
			defer dataNodeInstance.Close()

			request2 := GetFileArgs_c{}
			reply2 := GetFileReply_c{}
			request2.FileId = filelog.FileId
			request2.Offset = 0
			request2.SizeOfChunk = filelog.FileSize

			rpcErr = dataNodeInstance.Call("DataNode.GetFile_c", &request2, &reply2)
			if rpcErr != nil {
				//panic(rpcErr)
				fmt.Println(rpcErr)
				continue
				//return rpcErr
			}
			if reply2.Status == true {
				// err := os.Remove(dataNode.DataDirectory + request2.FileId)
				// if err != nil {
				// 	//panic(err)
				// 	//reply.Status = false
				// 	fmt.Println("Err in remove file")
				// 	continue
				// }
				file1, err2 := os.Create(dataNode.DataDirectory + request2.FileId)
				if err2 != nil {
					//panic(err)
					fmt.Println("Err in recreating file")
					continue
				}
				//defer file1.Close()
				_, err := file1.WriteAt(reply2.Data, request2.Offset)
				if err != nil {
					fmt.Println(err)
					continue
					//return nil
				}
				fmt.Println("Sending Done to Master")
				dataNodeInstance3, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
				if rpcErr != nil {
					//panic(rpcErr)
					fmt.Println(rpcErr)
					continue
				}
				//defer dataNodeInstance3.Close()

				request3 := DoneArgs{}
				request3.FileId = request2.FileId
				request3.FileSize = filelog.FileSize
				request3.Operation = "insert"
				var RepNodes []string
				RepNodes = append(RepNodes, Me)
				request3.ReplicationNodes = RepNodes
				request3.DoneTime = filelog.LastUpdated
				request3.NodeAddr = dataNode.Me
				// request3.BytesWritten = NoOfBytes
				reply3 := DoneReply{}

				rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
				if rpcErr != nil {
					//panic(rpcErr)
					fmt.Println(rpcErr)
					continue
				}
				fmt.Println("Sending Done to Master Successful", reply3.Status)
				dataNodeInstance3.Close()
				dataNodeInstance.Close()
				file1.Close()
			}
		}

		if filelog.Operation == "create" {
			//fmt.Println("Got Create file request from client")
			name, err2 := os.Create(dataNode.DataDirectory + filelog.FileId)
			fmt.Println(err2, name)
			if err2 != nil {
				//panic(err)
				fmt.Println(err2)
				continue
			}
			fmt.Println("Sending Done to Master")
			dataNodeInstance3, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
			if rpcErr != nil {
				//panic(rpcErr)
				fmt.Println(rpcErr)
				continue
			}
			//defer dataNodeInstance3.Close()

			request3 := DoneArgs{}
			request3.FileId = filelog.FileId
			request3.FileSize = filelog.FileSize
			request3.Operation = "create"
			var RepNodes []string
			RepNodes = append(RepNodes, Me)
			request3.ReplicationNodes = RepNodes
			request3.DoneTime = filelog.LastUpdated
			request3.NodeAddr = dataNode.Me
			// request3.BytesWritten = NoOfBytes
			reply3 := DoneReply{}

			rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
			if rpcErr != nil {
				//panic(rpcErr)
				fmt.Println(rpcErr)
				continue
			}
			fmt.Println("Sending Done to Master Successful", reply3.Status)
			dataNodeInstance3.Close()
			//dataNodeInstance.Close()
		}
		if filelog.Operation == "delete" {
			//fmt.Println("Got Create file request from client")

			err2 := os.Remove(dataNode.DataDirectory + filelog.FileId)
			if err2 != nil {
				//panic(err)
				fmt.Println(err2)
				//continue
			}
			fmt.Println("Sending Done to Master")
			dataNodeInstance3, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
			if rpcErr != nil {
				//panic(rpcErr)
				fmt.Println(rpcErr)
				//continue
			}
			defer dataNodeInstance3.Close()

			request3 := DoneArgs{}
			request3.FileId = filelog.FileId
			request3.FileSize = filelog.FileSize
			request3.Operation = "delete"
			var RepNodes []string
			RepNodes = append(RepNodes, Me)
			request3.ReplicationNodes = RepNodes
			request3.DoneTime = filelog.LastUpdated
			request3.NodeAddr = dataNode.Me

			// request3.BytesWritten = NoOfBytes
			reply3 := DoneReply{}
			reply3.Status = true
			rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
			if rpcErr != nil {
				//panic(rpcErr)
				fmt.Println(rpcErr)
				//continue
			}
			fmt.Println("Sending Done to Master Successful", reply3.Status)
			dataNodeInstance3.Close()
			//dataNodeInstance.Close()
		}

	}
	return nil
}

func (dataNode *DataNode) InsertFile_c(request *InsertFileArgs_c, reply *InsertFileReply_c) error {
	var NoOfBytes int
	if request.Offset == 0 {
		err := os.Remove(dataNode.DataDirectory + request.FileId)
		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		file1, err2 := os.Create(dataNode.DataDirectory + request.FileId)
		if err2 != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		defer file1.Close()
		_, err = file1.WriteAt(request.Data, request.Offset)
		if err != nil {
			fmt.Println(err)
			//return nil
		}

		fi, err := file1.Stat()
		if err != nil {
			//log.Fatal(err)
			fmt.Println(err)
			reply.Status = false
		}
		fmt.Println(fi.Size())
		NoOfBytes = int(fi.Size())
		reply.Status = true

	} else {
		file, err := os.OpenFile(dataNode.DataDirectory+request.FileId, os.O_RDWR, 0644)
		if err != nil {
			//panic(err)
			reply.Status = false
			fmt.Println("Open", err)
			return err
		}
		defer file.Close()

		//fileWriter := bufio.NewWriter(fileWriteHandler)

		_, err = file.WriteAt(request.Data, request.Offset)

		if err != nil {
			//panic(err)
			reply.Status = false
			fmt.Println("Write", err)
			return err
		}
		fi, err := file.Stat()
		if err != nil {
			//log.Fatal(err)
			fmt.Println("Stats", err)
			reply.Status = false
		}
		fmt.Println(fi.Size())
		NoOfBytes = int(fi.Size())
		reply.Status = true
	}

	//return dataNode.forwardForReplication(request, reply)

	err := dataNode.forwardForReplicationInsert(NoOfBytes, request, reply)
	if err != nil {
		reply.Status = false
		return nil
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

func (dataNode *DataNode) ForwardedForReplicationInsert(request *InsertFileArgs_c, reply *InsertFileReply_c) error {
	var NoOfBytes int
	if request.Offset == 0 {
		err := os.Remove(dataNode.DataDirectory + request.FileId)
		if err != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		file1, err2 := os.Create(dataNode.DataDirectory + request.FileId)
		if err2 != nil {
			//panic(err)
			reply.Status = false
			return err
		}
		_, err = file1.WriteAt(request.Data, request.Offset)
		if err != nil {
			fmt.Println(err)
			//return nil
		}

		fi, err := file1.Stat()
		if err != nil {
			//log.Fatal(err)
			fmt.Println(err)
			reply.Status = false
		}
		fmt.Println(fi.Size())
		NoOfBytes = int(fi.Size())
		reply.Status = true

	} else {
		file, err := os.Open(dataNode.DataDirectory + request.FileId)
		if err != nil {
			//panic(err)
			reply.Status = false
			fmt.Println(err)
			return err
		}
		defer file.Close()

		//fileWriter := bufio.NewWriter(fileWriteHandler)

		_, err = file.WriteAt(request.Data, request.Offset)

		if err != nil {
			//panic(err)
			reply.Status = false
			fmt.Println(err)
			return err
		}
		fi, err := file.Stat()
		if err != nil {
			//log.Fatal(err)
			fmt.Println(err)
			reply.Status = false
		}
		fmt.Println(fi.Size())
		NoOfBytes = int(fi.Size())
		reply.Status = true
	}
	if reply.Status == true {
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
		request3.FileSize = NoOfBytes
		request3.Operation = "insert"
		request3.DoneTime = time.Now()
		fmt.Println(request3.DoneTime)
		request3.NodeAddr = dataNode.Me
		// request3.BytesWritten = NoOfBytes
		reply3 := DoneReply{}
		reply.FileSize = int64(NoOfBytes)
		fmt.Println("filesize:", reply.FileSize)
		rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}
		fmt.Println("Sending Done to Master Successful", reply3.Status)

	}

	return nil

}

func (dataNode *DataNode) forwardForReplicationInsert(NoOfBytes int, request *InsertFileArgs_c, reply *InsertFileReply_c) error {

	fmt.Println("forwardForReplication")
	dataNodeInstance2, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
	if rpcErr != nil {
		//panic(rpcErr)
		//fmt.Println("sent request replication Nodes", rpcErr)
		//reply.Status = false
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
		//reply.Status = false
		return nil
	}
	fmt.Println("got replication Nodes")
	fmt.Println(len(reply2.ReplicationNodes))
	reply.Status = true
	for _, addr := range reply2.ReplicationNodes {

		if addr == Me {
			continue
		}
		dataNodeInstance, _ := rpc.DialHTTP("tcp", addr)

		// if rpcErr != nil {
		// 	//panic(rpcErr)
		// 	reply.Status = false
		// 	return rpcErr
		// }
		if dataNodeInstance != nil {
			defer dataNodeInstance.Close()
			err := dataNodeInstance.Call("DataNode.ForwardedForReplicationInsert", &request, &reply)
			if err != nil {
				return nil
			}
		}

	}
	fmt.Println(reply.Status)
	reply.FileSize = int64(NoOfBytes)
	fmt.Println(reply.FileSize)
	if reply.Status == true {
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
		request3.FileSize = NoOfBytes
		request3.Operation = "insert"
		request3.ReplicationNodes = reply2.ReplicationNodes
		request3.DoneTime = time.Now()
		fmt.Println(request3.DoneTime)
		request3.NodeAddr = dataNode.Me
		// request3.BytesWritten = NoOfBytes
		reply3 := DoneReply{}

		rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}
		fmt.Println("Sending Done to Master Successful", reply3.Status)

	}

	return nil
}

func (dataNode *DataNode) DeleteFile_m(request *DeleteFileArgs_m, reply *DeleteFileReply_m) error {
	fmt.Println("Got Delete file request from master")
	reply.Status = false
	// err := os.Remove(request.FileId)
	// if err != nil {
	// 	//panic(err)
	// 	reply.Status = false
	// 	return err
	// }
	err2 := os.Remove(dataNode.DataDirectory + request.FileId)
	fmt.Println(err2)
	if err2 != nil {
		//panic(err)
		reply.Status = false
		return nil
	}
	fmt.Println("File deleted")
	reply.Status = true
	err := dataNode.forwardForReplicationDelete(request, reply)
	if err != nil {
		reply.Status = false
		//return err
	}
	//reply.Status = true
	fmt.Println(reply.Status)
	return nil
}

func (dataNode *DataNode) ForwardedForReplicationDelete(request *DeleteFileArgs_m, reply *DeleteFileReply_m) error {

	_ = os.Remove(dataNode.DataDirectory + request.FileId)
	// if err != nil {
	// 	//panic(err)
	// 	// reply.Status = true
	// 	//return err

	// }
	// _, err2 := os.Create(dataNode.DataDirectory+request.FileId)
	// if err2 != nil {
	// 	//panic(err)
	// 	reply.Status = false
	// 	return err
	// }
	reply.Status = true
	fmt.Println("Sending Done to Master")
	dataNodeInstance3, rpcErr := rpc.DialHTTP("tcp", MasterAddr)
	if rpcErr != nil {
		//panic(rpcErr)
		reply.Status = false
		//return rpcErr
	}
	defer dataNodeInstance3.Close()

	request3 := DoneArgs{}
	request3.FileId = request.FileId
	request3.FileSize = 0
	request3.Operation = "delete"
	request3.DoneTime = time.Now()
	request3.NodeAddr = dataNode.Me
	// request3.BytesWritten = NoOfBytes
	//request3.doneTime = time.Now()
	reply3 := DoneReply{}
	//reply3.Status = true

	rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
	if rpcErr != nil {
		//panic(rpcErr)
		reply.Status = false
		//return rpcErr
	}
	fmt.Println("Sending Done to Master Successful", reply3.Status)

	return nil

}

func (dataNode *DataNode) forwardForReplicationDelete(request *DeleteFileArgs_m, reply *DeleteFileReply_m) error {
	fmt.Println("forwardForReplication")
	reply.Status = true
	for _, addr := range request.ReplicationNodes {
		fmt.Println(addr)
		if addr == Me {
			continue
		}

		dataNodeInstance, rpcErr := rpc.DialHTTP("tcp", addr)

		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			//return rpcErr
		}
		defer dataNodeInstance.Close()

		rpcErr = dataNodeInstance.Call("DataNode.ForwardedForReplicationDelete", &request, &reply)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			//break
			//return rpcErr
		}
	}
	fmt.Println(reply.Status)

	return nil
}

func (dataNode *DataNode) GetFile_c(request *GetFileArgs_c, reply *GetFileReply_c) error {
	// dataBytes, err := ioutil.ReadFile(dataNode.DataDirectory + request.BlockId)
	// //Check(err)
	// *reply = DataNodeData{Data: string(dataBytes)}
	// return nil

	file, err := os.OpenFile(dataNode.DataDirectory+request.FileId, os.O_RDWR, 0644)
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
	dataNode.ClientToken = request.UserToken
	// if dataNode.ClientToken == request.UserToken {
	reply.Status = true
	// } else {
	// reply.Status = false
	// }

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

func (dataNode *DataNode) CreateFile_c(request *CreateFileArgs_c, reply *CreateFileReply_c) error {
	fmt.Println("Got Create file request from client")
	// err := os.Remove(dataNode.DataDirectory+request.FileId)
	// if err != nil {
	// 	//panic(err)
	// 	reply.Status = false
	// 	return err
	// }
	reply.Status = true
	name, err2 := os.Create(dataNode.DataDirectory + request.FileId)
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

	fmt.Println(reply.Status)
	return nil
}

func (dataNode *DataNode) ForwardedForReplicationCreate(request *InsertFileArgs_c, reply *InsertFileReply_c) error {
	fmt.Println("I was called for create")

	// err := os.Remove(dataNode.DataDirectory + request.FileId)
	// if err != nil {
	// 	//panic(err)
	// 	reply.Status = false
	// 	return err
	// }
	_, err2 := os.Create(dataNode.DataDirectory + request.FileId)
	if err2 != nil {
		//panic(err)
		reply.Status = false
		//return err
	}
	reply.Status = true
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
	request3.DoneTime = time.Now()
	request3.NodeAddr = dataNode.Me
	// request3.BytesWritten = NoOfBytes
	//request3.doneTime = time.Now()
	reply3 := DoneReply{}

	rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
	if rpcErr != nil {
		//panic(rpcErr)
		reply.Status = false
		return rpcErr
	}
	fmt.Println("Sending Done to Master Successful", reply3.Status)

	return nil

}

func (dataNode *DataNode) forwardForReplicationCreate(request *CreateFileArgs_c, reply *CreateFileReply_c) error {
	fmt.Println("forwardForReplication")
	dataNodeInstance2, _ := rpc.DialHTTP("tcp", MasterAddr)
	// if rpcErr != nil {
	// 	//panic(rpcErr)
	// 	fmt.Println("sent request replication Nodes", rpcErr)
	// 	//reply.Status = false
	// 	return nil
	// }
	defer dataNodeInstance2.Close()

	request2 := GetReplicationNodesArgs{}
	reply2 := GetReplicationNodesReply{}
	// request2:=*GetReplicationNodes_Args
	// reply2:=*GetReplicationNodes_Reply

	_ = dataNodeInstance2.Call("Coordinator.GetReplicationNodes", &request2, &reply2)
	// if rpcErr3 != nil {
	// 	//panic(rpcErr)
	// 	//reply2.Status = false
	// 	//return rpcErr
	// 	//reply.Status = false
	// 	return nil
	// }
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
			fmt.Println(rpcErr)
			//return rpcErr
			continue
		}
		defer dataNodeInstance.Close()

		rpcErr = dataNodeInstance.Call("DataNode.ForwardedForReplicationCreate", &request, &reply)
		if rpcErr != nil {
			//panic(rpcErr)
			fmt.Println(rpcErr)
			//return rpcErr
			continue
		}
	}
	fmt.Println(reply.Status)
	if reply.Status == true {
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
		request3.DoneTime = time.Now()
		request3.NodeAddr = dataNode.Me
		// request3.BytesWritten = NoOfBytes
		//request3.doneTime = time.Now()
		reply3 := DoneReply{}

		rpcErr = dataNodeInstance3.Call("Coordinator.Done", &request3, &reply3)
		if rpcErr != nil {
			//panic(rpcErr)
			reply.Status = false
			return rpcErr
		}
		fmt.Println("Sending Done to Master Successful", reply3.Status)

	}

	return nil
}
