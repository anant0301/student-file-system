package main

import (
	"fmt"
	"time"
	"os"
	"strconv"
)

func main() {
	c := InitServer()
	fmt.Println("Server has started")
	c.DataDirectory=os.Args[2]
	//dataDirectory:=os(mkdir)
	//fmt.Printf("%v\n", os.Args[1])
	//fmt.Printf("%v\n", os.Args[2])
	var host string=os.Args[1]
	var port int
	port,_=strconv.Atoi(os.Args[2])
	fmt.Println(host,port)
	fmt.Printf("%v\n", c.DataDirectory)
	for {
		c.Ping(host,port)
		time.Sleep(time.Second * 5)
	}
}
