package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/alexandear/websocket-pubsub/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := cmd.Exec(os.Args); err != nil {
		log.Fatal(err)
	}
}
