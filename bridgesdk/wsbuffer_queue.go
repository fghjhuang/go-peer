package bridgesdk

import (
	"github.com/gorilla/websocket"
	"sync"
)

type WebsocketObject struct {
	opcode int // 操作类型
	did    string
	conn   *websocket.Conn
}

//缓存
var wsBuffer map[string]*websocket.Conn

// ws通道
var wsQueueChannel = make(chan WebsocketObject, 1)
var m *sync.RWMutex

//queue
var startWsQueue = false

func init() {
	m = new(sync.RWMutex)
	wsBuffer = make(map[string]*websocket.Conn)
	readWSThread()
}

func putWsBuffer(did WebsocketObject) {
	wsQueueChannel <- did
}

func stopQueue() {
	startWsQueue = false
}

func readWSThread() {
	//fmt.Print("WS queue start read\n")
	startWsQueue = true
	go func() {
		for {
			select {
			case poll := <-wsQueueChannel:
				m.Lock()
				switch poll.opcode {
				case 1: // add
					wsBuffer[poll.did] = poll.conn // 分配缓存空间
					break
				case 2: // delete
					delete(wsBuffer, poll.did)
					break
				}
				m.Unlock()
			}
			if !startWsQueue {
				break
			}
		}
	}()
}
