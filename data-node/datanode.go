package main

import (
	"fmt"
	"net/rpc"
	"os"
	"sync"

	"gopkg.in/ini.v1" // ini parser
)

type DNServer struct {
	mServerAddr  []string
	mServer      []*rpc.Client
	is_connected bool
	mu           sync.Mutex
}

func (dn *DNServer) connect(hostAddr string) bool {
	dn.mu.Lock()
	dn.mServerAddr = append(dn.mServerAddr, hostAddr)
	dn.mu.Unlock()
	client, err := rpc.DialHTTP("tcp", hostAddr)
	dn.mu.Lock()
	dn.mServer = append(dn.mServer, client)
	dn.mu.Unlock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in connecting to %s")
		return false
	}
}

func InitServer() *DNServer {
	cfg, err := ini.Load(".ini")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in parsing .ini file\n")
		os.Exit(1)
	}
	dn := &DNServer{}
	var hosts []string
	for i := 1; i <= 3; i++ {
		hosts = append(hosts, cfg.Section("rpc").Key(fmt.Sprintf("master-%d", i)).String())
	}

	return dn
}
