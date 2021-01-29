package server

import (
	"time"
)

type Message struct {
	Data MessageData
}

type MessageData interface{}

type BroadcastData struct {
	Time time.Time
}

type UnicastData struct {
	ClientID string
}

type ResponseMessage interface{}

type ResponseBroadcast struct {
	ClientID string
	Time     time.Time
}

type ResponseUnicast struct {
	NumConnections int
}
