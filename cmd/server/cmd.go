package server

import (
	"time"

	flag "github.com/spf13/pflag"

	"github.com/alexandear/websocket-pubsub/internal/server"
)

const (
	defaultBroadcast = 100 * time.Millisecond
)

func Exec() error {
	addr := flag.String("addr", ":8080", "http service address")
	broadcast := flag.Duration("broadcast", defaultBroadcast, "broadcast frequency")

	flag.Parse()

	a := server.New(*addr, *broadcast)

	return a.Run()
}
