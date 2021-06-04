package beans

import "encoding/json"

const (
	REGISTER = "register"
	CONNECT  = "connect"
)

type PeerSignal struct {
	DID    string `json:"did"`
	Type   string `json:"type"`
	TarDID string `json:"tar_did"`
}

func (peer *PeerSignal) ToString() string {
	result, _ := json.Marshal(peer)
	return string(result)
}