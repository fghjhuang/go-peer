package main

import (
	"gopeer/serversdk"
	"time"
)

func main() {
	server := serversdk.PeerServer{}
	server.CreateServer()
	for {
		time.Sleep(time.Second * 20)
	}
}
