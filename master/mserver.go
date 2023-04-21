package main

//
// start the master node process, which is implemented
//
import (
	"fmt"
	"time"
)

func main() {
	c := InitServer()
	fmt.Println("Server has started")
	for {
		time.Sleep(time.Second * 100)
		fmt.Println("Data Nodes Available", c.mcon.getServers())
	}
}
