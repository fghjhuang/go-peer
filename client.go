package main

import (
	"gopeer/clientsdk"
	"time"
)

func main() {
	client := clientsdk.PeerClient{}
	client.CreateClient()
	for {
		time.Sleep(time.Second * 20)
	}
}
