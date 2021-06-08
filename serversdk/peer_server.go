package serversdk

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"gopeer/global"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//https://pkg.go.dev/modernc.org/sqlite

/*
* 服务端SDK
 */

type PeerServer struct {
}

var tag string
var signalAddress = ""
var localAddress = ":9595"
var udpChannel = make(chan global.PeerSignal)
var did = ""

func (s *PeerServer) CreateServer() {
	// p2p服务器地址
	signalAddress = os.Args[2]

	//本地端口
	if len(os.Args) > 3 {
		localAddress = os.Args[3]
	}

	//设备ID
	did = os.Args[4]

	fmt.Println(fmt.Sprintf("server address is:%s,local port is:%s,device id is:%s", signalAddress, localAddress, did))
	go register()
	go udpHoleTask()
}

func register() {
	//ws连接信令服务器
	var singaladdr = flag.String("addr", signalAddress, "ws service address")
	u := url.URL{Scheme: "ws", Host: *singaladdr, Path: "/signal"}
	heads := http.Header{}
	heads["Authorization"] = []string{"test"}
	heads["username"] = []string{"test"}
	heads["password"] = []string{"test"}
	heads["did"] = []string{did}

	webcon, _, err := websocket.DefaultDialer.Dial(u.String(), heads)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer webcon.Close()
	peerRequest := global.PeerSignal{}

	done := make(chan bool)

	//循环读取
	go func() {
		for {
			_, message, err := webcon.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				done <- true
			}
			err = json.Unmarshal(message, &peerRequest)
			if err != nil {
				fmt.Println("analyse data error", err.Error())
				continue
			}
			udpChannel <- peerRequest //直接在udp线程里面处理
			log.Printf("ws recv: %s", message)
		}
	}()

	for {
		select {
		case <-done:
			return
		}
	}
}

var udpHoleList map[string]string

func udpHoleTask() {
	udpHoleList = make(map[string]string)
	signalServer, _ := net.ResolveUDPAddr("udp", signalAddress)
	local, _ := net.ResolveUDPAddr("udp", localAddress)
	conn, _ := net.ListenUDP("udp", local)
	peerFeedback := global.PeerSignal{}
	sayback := make(chan string)
	//udp接收
	go func() {
		buffer := make([]byte, 256)
		for {
			cnt, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("[ERROR]", err)
				continue
			}
			fmt.Println("[udp incoming]:" + string(buffer[:cnt]))
			sayback <- "[from server]: Hello!"
		}
	}()

	for {
		select {
		case peerdata := <-udpChannel: //接收peer请求
			switch peerdata.Type {
			case global.CONNECT: //请求连接，发送udp到服务器进行穿透
				peerFeedback.Type = global.FEEDBACK
				peerFeedback.DID = did
				peerFeedback.TarDID = peerdata.DID
				udpHoleList[peerdata.DID] = peerdata.DIDIP //保存对方的DID和地址
				data, _ := json.Marshal(peerFeedback)
				conn.WriteTo(data, signalServer)
				fmt.Println("[udp out]:" + string(data))
				go func() {
					addr, _ := net.ResolveUDPAddr("udp", peerdata.DIDIP)
					for {
						conn.WriteTo([]byte("ping"), addr)
						fmt.Println("ping state:" + peerdata.DIDIP)
						time.Sleep(time.Second * 3)
					}
				}()
				break
			}
			break
		}
	}
}

func listen(conn *net.UDPConn, local string) {
	for {
		fmt.Println("listening")
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("[ERROR]", err)
			continue
		}

		fmt.Println("[INCOMING]", string(buffer[0:bytesRead]))
		// 如果是Hello开头的就直接再等待数据
		if string(buffer[0:bytesRead]) == "Hello!" {
			continue
		}

		//如果不是普通文本就是告知目标地址和端口
		for _, a := range strings.Split(string(buffer[0:bytesRead]), ",") {
			if a != local { // 不等于本地地址才进行通信
				go chatter(conn, a)
			}
		}
	}
}

func chatter(conn *net.UDPConn, remote string) {
	addr, _ := net.ResolveUDPAddr("udp", remote)
	for {
		conn.WriteTo([]byte("Hello!"), addr)
		fmt.Println("sent Hello! to ", remote)
		time.Sleep(5 * time.Second)
	}
}
