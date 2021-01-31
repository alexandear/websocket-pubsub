package server

import (
	"time"
)

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
