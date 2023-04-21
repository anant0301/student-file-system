package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"gopkg.in/ini.v1"
)

type Coordinator struct {
	mcon MongoConnector
}

func (c *Coordinator) init_mongo() {
	c.mcon = MongoConnector{}
	c.mcon.connect()
}

func (c *Coordinator) server(host string, port int) {
	rpc.Register(c)
	rpc.HandleHTTP()
	sockname := fmt.Sprintf("%s:%d", host, port)
	l, e := net.Listen("tcp", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("RPC Conenction", host, port)
	go http.Serve(l, nil)
}

func InitServer() *Coordinator {
	cfg, err := ini.Load(".ini")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in parsing .ini file\n")
		os.Exit(1)
	}
	var host string = cfg.Section("rpc").Key("host").String()
	var port int = cfg.Section("rpc").Key("port").MustInt()
	c := Coordinator{}
	c.server(host, port)
	c.init_mongo()
	return &c
}

func (c *Coordinator) DialDataNode(serverAddr string, rpcCall string, args interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", serverAddr)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	ok := client.Call(rpcCall, args, reply)
	return ok
}
