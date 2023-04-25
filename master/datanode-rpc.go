package main

import (
	"log"
	"time"
)

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
	FileId           string
	FileSize         int64
	Operation        string
	ReplicationNodes []string
	doneTime         time.Time
}

const (
	/* File operations */
	INSERT = "insert"
	CREATE = "create"
	DELETE = "delete"
)

type DoneReply struct {
	Status bool
}

type PingArgs struct {
	Addr string
	// FreeSpace int
}

type FileLog struct {
	Addr        string
	Operation   string
	FileId      string
	FileSize    int
	LastUpdated time.Time
}

type PingReply struct {
	Logs []FileLog
}

func (c *Coordinator) Ping(args *PingArgs, reply *PingReply) error {
	log.Println("Pinging Master", args)
	c.mcon.updatePing(args.Addr)
	logs := c.mcon.getInconsistentLogs(args.Addr)
	for _, log := range logs {
		reply.Logs = append(reply.Logs, FileLog{
			Addr:        log.address,
			Operation:   log.operation,
			FileId:      log.fileId,
			FileSize:    log.fileSize,
			LastUpdated: log.lastUpdated,
		})
	}
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
	for _, node := range args.ReplicationNodes {
		// c.mcon.updateServerLastOperation(node, args.Operation)
		c.mcon.updateLogsNode(node, args.FileId, args.Operation, args.doneTime)
	}
	return nil
}
