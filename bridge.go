package main

import "gopeer/bridgesdk"

func main() {
	bridge := bridgesdk.Bridge{Port: 12501}

	bridge.CreateBridgeServer()

}
