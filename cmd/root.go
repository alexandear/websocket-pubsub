package cmd

import (
	"errors"
	"fmt"

	"github.com/alexandear/websocket-pubsub/cmd/client"
	"github.com/alexandear/websocket-pubsub/cmd/server"
)

var ErrBadCommand = errors.New("wrong command")

const (
	commandClient = "client"
	commandServer = "server"
)

func Exec(args []string) error {
	if len(args) <= 1 {
		return fmt.Errorf("missing command: %w", ErrBadCommand)
	}

	switch command := args[1]; command {
	case commandClient:
		return client.Exec()
	case commandServer:
		return server.Exec()
	default:
		return fmt.Errorf("unknown command %s: %w", command, ErrBadCommand)
	}
}
