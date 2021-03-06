package operation

import (
	"github.com/alexandear/websocket-pubsub/internal/pkg/command"
)

type ReqCommand struct {
	Command command.Type `json:"command"`
}

type Resp interface{}

type RespBroadcast struct {
	ClientID  string `json:"client_id"`
	Timestamp int    `json:"timestamp"`
}

type RespNumConnections struct {
	NumConnections int `json:"num_connections"`
}
