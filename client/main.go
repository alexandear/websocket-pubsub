package main

import (
	"context"
	"flag"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	addr := flag.String("addr", "localhost:8080", "http server address")
	clients := flag.Int("clients", 5000, "number of clients")

	flag.Parse()

	app := NewApp(*addr, *clients)
	app.Run(context.Background())
}
