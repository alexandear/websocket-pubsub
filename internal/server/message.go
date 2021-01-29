package server

import (
	"time"
)

type CastType int

const (
	Broadcast CastType = iota
	Unicast
)

type MessageData int

const (
	MessageDataTime MessageData = iota
	MessageDataNumConn
)

type Data interface {
	Type() MessageData
}

type Message struct {
	CastType CastType
	Data     Data
}

type BroadcastData struct {
	Time time.Time
}

func (d *BroadcastData) Type() MessageData {
	return MessageDataTime
}

type UnicastData struct {
	ClientID string
}

func (d *UnicastData) Type() MessageData {
	return MessageDataNumConn
}

type ResponseMessage interface{}

type ResponseBroadcast struct {
	ResponseMessage

	ClientID string
	Time     time.Time
}

type ResponseUnicast struct {
	ResponseMessage

	NumConnections int
}
