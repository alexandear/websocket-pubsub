package server

import (
	"time"
)

type Communication int

const (
	CommunicationBroadcast Communication = iota
	CommunicationUnicast
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
	Communication Communication
	Data          Data
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
