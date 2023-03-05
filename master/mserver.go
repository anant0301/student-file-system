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
	fmt.Println("This is done now")

	time.Sleep(time.Second)
}
