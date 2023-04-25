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
		// args := PingArgs{
		// 	Addr: "10.0.60.1000:9000",
		// }
		// reply := PingReply{}
		// c.Ping(&args, &reply)
		// fmt.Println(reply)
		time.Sleep(time.Second * 100)
		fmt.Println("Data Nodes Available", c.mcon.getServers())
	}
}
