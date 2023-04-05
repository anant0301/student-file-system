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
	fmt.Printf("%v\n", c.serverlist)
	for {
		time.Sleep(time.Second * 10)
	}
}
