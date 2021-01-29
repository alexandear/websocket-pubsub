package client

import (
	"context"

	flag "github.com/spf13/pflag"

	"github.com/alexandear/websocket-pubsub/internal/client"
)

func Exec() error {
	addr := flag.String("addr", "localhost:8080", "http server address")
	clients := flag.Int("clients", 5000, "number of clients")

	flag.Parse()

	app := client.NewApp(*addr, *clients)
	app.Run(context.Background())

	return nil
}
