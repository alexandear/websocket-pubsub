package main

import (
	"context"
	"flag"
	"log"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "http server address")

	flag.Parse()

	app := NewApp(*addr)
	if err := app.Run(context.Background()); err != nil {
		log.Fatalf("run failed: %v", err)
	}
}
