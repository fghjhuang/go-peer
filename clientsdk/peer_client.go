package clientsdk

import (
	"fmt"
	"gopeer/global"
	"net"
	"os"
	"strings"
	"time"
)

/*
* 客户端的SDK
 */
type PeerClient struct {
}

func (s *PeerClient) CreateClient() {
	register()
}

func register() {
	// p2p服务器地址
	signalAddress := os.Args[2]

	//本地端口
	localAddress := ":9595" // default port
	if len(os.Args) > 3 {
		localAddress = os.Args[3]
	}

	remote, _ := net.ResolveUDPAddr("udp", signalAddress)
	local, _ := net.ResolveUDPAddr("udp", localAddress)
	conn, _ := net.ListenUDP("udp", local)

	registerinfo := global.PeerSignal{
		DID:    os.Args[4],
		Type:   "connect",
		TarDID: os.Args[5],
	}

	// 注册云端服务
	go func() {
		bytesWritten, err := conn.WriteTo([]byte(registerinfo.ToString()), remote)
		if err != nil {
			panic(err)
		}

		fmt.Println(bytesWritten, " bytes written")
	}()

	listen(conn, local.String())
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
