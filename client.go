package main

import "gopeer/clientsdk"

func main() {
	client := clientsdk.PeerClient{}
	client.CreateClient()
}
