package operation

import (
	"encoding/json"
	"time"

	"github.com/alexandear/websocket-pubsub/command"
)

type ReqCommand struct {
	Command command.Type `json:"command"`
}

type RespBroadcast struct {
	ClientID  string
	Timestamp time.Time
}

func (r *RespBroadcast) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ClientID  string `json:"client_id"`
		Timestamp string `json:"timestamp"`
	}{
		ClientID:  r.ClientID,
		Timestamp: r.Timestamp.Format(time.RFC3339),
	})
}
