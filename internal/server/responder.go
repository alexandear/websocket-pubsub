package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alexandear/websocket-pubsub/internal/pkg/operation"
)

type Responder struct{}

func (r *Responder) Bytes(message Data) ([]byte, error) {
	switch message.Type() {
	case MessageDataNumConn:
		d, ok := message.(*UnicastData)
		if !ok {
			return nil, errors.New("wrong unicast data")
		}

		return r.numConnections(d.NumConnections)
	case MessageDataTime:
		d, ok := message.(*BroadcastData)
		if !ok {
			return nil, errors.New("wrong broadcast data")
		}

		return r.broadcast(d.ClientID, d.Time)
	default:
		return nil, errors.New("unknown message data type")
	}
}

func (r *Responder) numConnections(numConnections int) ([]byte, error) {
	b, err := json.Marshal(&operation.RespNumConnections{
		NumConnections: numConnections,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal RespNumConnections failed: %w", err)
	}

	return b, nil
}

func (r *Responder) broadcast(clientID string, t time.Time) ([]byte, error) {
	b, err := json.Marshal(&operation.RespBroadcast{
		ClientID:  clientID,
		Timestamp: int(t.Unix()),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal RespBroadcast failed: %w", err)
	}

	return b, nil
}
