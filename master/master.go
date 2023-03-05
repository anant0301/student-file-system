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
	mcon  MongoConnector
	reqid int
}

func (c *Coordinator) init_mongo() {
	mcon := MongoConnector{}
	mcon.connect()
}

func (c *Coordinator) server(host string, port int) {
	rpc.Register(c)
	rpc.HandleHTTP()
	sockname := fmt.Sprintf("%s:%d", host, port)
	l, e := net.Listen("tcp", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
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
	fmt.Println("RPC Conenction", host, port)
	c := Coordinator{}
	c.server(host, port)
	c.init_mongo()
	return &c
}
