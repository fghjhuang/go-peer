package bridgesdk

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"gopeer/global"
	"log"
	"net"
	"net/http"
)

type Bridge struct {
	Port int
}

/*
* 创建穿透服务
 */

func (b *Bridge) CreateBridgeServer() {
	if b.Port == 0 {
		return
	}
	fmt.Println("start udp hole server")
	go udpBridge(b.Port)
	go createSignalServer()
	go createHttpServer()
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
	peerRequest := global.PeerSignal{}

	for {
		bytesRead, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}
		incoming := string(buffer[0:bytesRead])
		fmt.Println("[udp INCOMING]", incoming, "from ip:port is ", remoteAddr.String())
		err = json.Unmarshal([]byte(incoming), &peerRequest)
		if err != nil {
			fmt.Println("analyse data error", err.Error())
			continue
		}

		_, _ = conn.WriteTo([]byte("udp"), remoteAddr)

		switch peerRequest.Type {
		case global.REGISTER: //注册
			udpHoleList[peerRequest.DID] = remoteAddr.String()
			break

		case global.FEEDBACK:
			//ws转发到服务端
			m.RLock()
			value, ok := wsBuffer[peerRequest.TarDID]
			m.RUnlock()

			if ok {
				fmt.Println("find tar device: " + peerRequest.TarDID)
				peerRequest.DIDIP = remoteAddr.String() //加上端口信息转发
				data, err := json.Marshal(peerRequest)
				err = value.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					fmt.Println("write error")
				}
			} else {
				fmt.Println("not find device: " + peerRequest.TarDID)
			}
			break
		case global.CONNECT: //客户端请求连接服务端
			//ws转发到服务端
			m.RLock()
			value, ok := wsBuffer[peerRequest.TarDID]
			m.RUnlock()

			if ok {
				fmt.Println("find tar device: " + peerRequest.TarDID)
				peerRequest.DIDIP = remoteAddr.String() //加上端口信息转发
				data, err := json.Marshal(peerRequest)
				err = value.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					fmt.Println("write error")
				}
			} else {
				fmt.Println("not find device: " + peerRequest.TarDID)
			}
			break

		}

	}
}

/*--------------------------------------------------分割线 websocket信令交互---------------------------------------------*/

var addr = flag.String("addr", ":12501", "ws service address")

//默认先不认证
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		auth := r.Header.Get("Authorization")
		username := r.Header.Get("username")
		password := r.Header.Get("password")
		if auth != "" && username != "" && password != "" {
			// TODO 这里做认证
			return true
		} else {
			return true
		}
		// return true
	},
} // use default options

func signal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	fmt.Println("[new ws connection]")

	did := r.Header.Get("did")

	if did == "" {
		fmt.Println("new connection not find did info, will be return")
		return
	}

	defer func() {
		conn.Close()
		putWsBuffer(WebsocketObject{
			opcode: 2,
			did:    did,
			conn:   conn,
		})
	}()

	//注册连接
	putWsBuffer(WebsocketObject{
		opcode: 1,
		did:    did,
		conn:   conn,
	})

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = conn.WriteMessage(mt, message)
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
