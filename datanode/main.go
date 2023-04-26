package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

var Me string

func main() {
	c := InitServer()
	fmt.Println("Server has started")
	c.DataDirectory = "data-" + os.Args[2] + "/"

	if _, err := os.Stat(c.DataDirectory); os.IsNotExist(err) {
		os.Mkdir(c.DataDirectory, 0777)
	}
	//dataDirectory:=os(mkdir)
	//fmt.Printf("%v\n", os.Args[1])
	//fmt.Printf("%v\n", os.Args[2])
	var host string = os.Args[1]
	var port int
	port, _ = strconv.Atoi(os.Args[2])
	Me = host + ":" + strconv.Itoa(port)
	c.Me = Me

	fmt.Println(host, port)
	fmt.Printf("%v\n", c.DataDirectory)
	for {
		c.Ping(host, port)
		time.Sleep(time.Second * 10)
	}
}
