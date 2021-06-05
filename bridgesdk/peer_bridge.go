package bridgesdk

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"gopeer/beans"
	"log"
	"net"
	"net/http"
)

type Bridge struct {
	Port int
}

/*
* 创建中转服务
 */

func (b *Bridge) CreateBridgeServer() {
	if b.Port == 0 {
		return
	}
	fmt.Println("start udp hole server")
	go udpBridge(b.Port)
	go createSignalServer()
}

/*--------------------------------------------------分割线 udp打洞------------------------------------------------------*/
var udpHoleList map[string]string

// udp 穿透
func udpBridge(port int) {
	udpHoleList = make(map[string]string, 0)
	localAddress := fmt.Sprintf(":%d", port)

	udpserverAddr, err := net.ResolveUDPAddr("udp", localAddress)

	if err != nil {
		fmt.Println("create udp hole server error", err.Error())
		return
	}

	// 监听服务器的udp端口
	conn, _ := net.ListenUDP("udp", udpserverAddr)
	buffer := make([]byte, 128)
	peerRequest := beans.PeerSignal{}
	for {
		bytesRead, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}
		incoming := string(buffer[0:bytesRead])
		fmt.Println("[INCOMING]", incoming)
		err = json.Unmarshal([]byte(incoming), &peerRequest)
		if err != nil {
			fmt.Println("analyse data error", err.Error())
			continue
		}

		switch peerRequest.Type {
		case beans.REGISTER: //注册
			udpHoleList[peerRequest.DID] = remoteAddr.String()
			break

		case beans.CONNECT: //连接
			_, ok := udpHoleList[peerRequest.TarDID]
			if ok {
				udpHoleList[peerRequest.DID] = remoteAddr.String()
				// 通知目标设备本地址
				r, _ := net.ResolveUDPAddr("udp", udpHoleList[peerRequest.TarDID])
				conn.WriteTo([]byte(remoteAddr.String()), r)
				fmt.Printf("[INFO] Responded to %s with %s\n", udpHoleList[peerRequest.TarDID], string(remoteAddr.String()))

				//通知本设备目标设备地址
				r, _ = net.ResolveUDPAddr("udp", udpHoleList[peerRequest.DID])
				conn.WriteTo([]byte(udpHoleList[peerRequest.TarDID]), r)
				fmt.Printf("[INFO] Responded to %s with %s\n", udpHoleList[peerRequest.DID], string(udpHoleList[peerRequest.TarDID]))
			}
			break
		}

	}
}

/*--------------------------------------------------分割线 websocket信令交互---------------------------------------------*/
var addr = flag.String("addr", ":8698", "http service address")

//默认先不认证
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		auth := r.Header.Get("Authorization")
		username := r.Header.Get("username")
		if auth != "" && username != "" {
			return true
		} else {
			return true
		}
		// return true
	},
} // use default options

func signal(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

/*
* 创建信令服务
 */
func createSignalServer() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/signal", signal)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
