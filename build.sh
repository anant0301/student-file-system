#!/bin/bash

# building master servers
cd master
go build mserver.go master.go rpc.go structs.go mongo.go client-rpc.go datanode-rpc.go


# building client servers
cd client
go mod download github.com/hanwen/go-fuse/v2
go build main.go rpc.go 
./main <master-addr> <mount-dir> <root-parent> <root-name>


# building data node servers
cd datanode
go get github.com/hanwen/go-fuse/v2/f
go build main.go rpc.go 
go build main.go rpc.go 
git push
cd ../datanode/
go build datanode-rpc.go datanode.go main.go rpc.go 
./datanode-rpc  <data-node-ip> <data-node-port>