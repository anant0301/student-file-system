package main

import "log"

type DataNode struct {
	Addr    string
	IsAlive bool
}

type InsertFileArgs_m struct {
	UserToken string
}

type InsertFileReply_m struct {
	Status bool
}

type CreateFileArgs_m struct {
	UserToken string
}

type CreateFileReply_m struct {
	Status bool
}

type DeleteFileArgs_m struct {
	UserToken        string
	FileId           string
	ReplicationNodes []string
}

type DeleteFileReply_m struct {
	Status bool
}

type GetReplicationNodesArgs struct{}

type GetReplicationNodesReply struct {
	ReplicationNodes []string
}

type DoneArgs struct {
	FileId    string
	FileSize  int64
	Operation string
}

type DoneReply struct {
	Status bool
}

func (c *Coordinator) Ping(args *PingArgs, reply *PingReply) error {
	log.Println("Pinging Master", args)
	reply.Status = c.mcon.updateDataNode(args.Addr, true) > 0
	return nil
}

func (c *Coordinator) GetReplicationNodes(args *GetReplicationNodesArgs, reply *GetReplicationNodesReply) error {
	log.Println("GetReplicationNodes")
	dnodes := c.mcon.getServers()
	for _, dnode := range dnodes {
		if dnode.IsAlive == true {
			reply.ReplicationNodes = append(reply.ReplicationNodes, dnode.Addr)
		}
	}
	return nil
}

func (c *Coordinator) Done(args *DoneArgs, reply *DoneReply) error {
	log.Println("DoneReq:", args.Operation, "on file", args.FileId, "file size is now", args.FileSize)
	reply.Status = c.mcon.updateFileDone(args.FileId, args.FileSize) > 0
	return nil
}
