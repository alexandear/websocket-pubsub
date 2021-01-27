package operation

import (
	"github.com/alexandear/websocket-pubsub/command"
)

type ReqCommand struct {
	Command command.Type `json:"command"`
}

type RespBroadcast struct {
	ClientID  string `json:"client_id"`
	Timestamp string `json:"timestamp"`
}

type RespNumConnections struct {
	NumConnections int `json:"num_connections"`
}
