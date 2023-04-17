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

type DataNode struct {
	Host        string
	Id          int
	FreeSpace   int
	Heartbeat   int
	ClientToken string
}

func (d *DataNode) server(host string, port int) {
	// rpc.Ping(c)
	rpc.HandleHTTP()
	sockname := fmt.Sprintf("%s:%d", host, port)
	l, e := net.Listen("tcp", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func InitServer() *DataNode {
	cfg, err := ini.Load(".ini")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in parsing .ini file\n")
		os.Exit(1)
	}
	var host string = cfg.Section("rpc").Key("host").String()
	var port int = cfg.Section("rpc").Key("port").MustInt()
	fmt.Println("RPC Conenction", host, port)
	c := DataNode{}
	c.server(host, port)
	return &c
}
