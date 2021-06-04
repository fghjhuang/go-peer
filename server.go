package main

import (
	"gopeer/serversdk"
)

func main() {
	server := serversdk.PeerServer{}
	server.CreateServer()
}
