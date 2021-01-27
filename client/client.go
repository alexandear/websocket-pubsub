package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/command"
	"github.com/alexandear/websocket-pubsub/operation"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	uhttp := url.URL{
		Scheme: "http",
		Host:   *addr,
		Path:   "/pubsub",
	}

	log.Printf("connecting to %s", uhttp.String())

	bs, err := json.Marshal(&operation.ReqCommand{
		Command: command.Subscribe,
	})
	if err != nil {
		log.Fatal("marshal:", err)
	}

	req, err := http.NewRequest(http.MethodGet, uhttp.String(), bytes.NewReader(bs))
	if err != nil {
		log.Fatal("failed to create request:", err)
	}

	clientID := uuid.New().String()
	clientID = "3d5ca28a-4f22-4fea-b2c8-23de1ac72ab9"
	req.Header.Set("X-Client-Id", clientID)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("failed to do request:", err)
	}

	loc, err := resp.Location()
	if err == http.ErrNoLocation {
		log.Println("nothing to do")

		return
	}
	if err != nil {
		log.Fatal("failed to get location:", err)
	}

	log.Printf("redirecting to %s", loc.String())

	c, _, err := websocket.DefaultDialer.Dial(loc.String(), http.Header{
		"X-Client-Id": []string{clientID},
	})
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			b, err := json.Marshal(&operation.ReqCommand{
				Command: command.Unsubscribe,
			})
			if err != nil {
				log.Println("marshal:", err)
				return
			}

			if err := c.WriteMessage(websocket.BinaryMessage, b); err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			if err := c.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
