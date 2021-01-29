package server

import (
	"time"
)

type CastType int

const (
	Broadcast CastType = iota
	Unicast
)

type Message struct {
	CastType CastType
	Data     CastData
}

type CastData interface{}

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
