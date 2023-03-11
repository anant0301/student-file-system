package main

//
// start the master node process, which is implemented
//
import (
	"fmt"
	"time"
)

func main() {
	InitServer()
	fmt.Println("Server has started")
	time.Sleep(time.Second * 10)
}
