package main

import (
	"gopeer/bridgesdk"
	"time"
)

func main() {
	bridge := bridgesdk.Bridge{Port: 12501}

	bridge.CreateBridgeServer()

	for {
		time.Sleep(time.Second * 20)
	}
}
