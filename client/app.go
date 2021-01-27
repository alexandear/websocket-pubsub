package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"

	"github.com/alexandear/websocket-pubsub/command"
	"github.com/alexandear/websocket-pubsub/operation"
)

const (
	sendCommandsAfter        = 5 * time.Second
	timeoutBeforeUnsubscribe = 2 * time.Second

	httpClientTimeout = 2 * time.Second
)

type App struct {
	server string

	client *http.Client
	conn   *websocket.Conn

	done      chan struct{}
	interrupt chan os.Signal
}

func NewApp(server string) *App {
	app := &App{
		server: server,
		client: &http.Client{
			Timeout:       httpClientTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		},

		done:      make(chan struct{}),
		interrupt: make(chan os.Signal, 1),
	}

	signal.Notify(app.interrupt, os.Interrupt)

	return app
}

func (a *App) Run(ctx context.Context) error {
	redirect, err := a.subscribeRedirect(ctx, a.server, "/pubsub")
	if err != nil {
		return fmt.Errorf("subsribe failed: %w", err)
	}

	// nolint:bodyclose // does not need to be closed
	c, _, err := websocket.DefaultDialer.DialContext(ctx, redirect, nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	a.conn = c

	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("close failed: %v", err)
		}
	}()

	go a.readWs()

	a.writeWs()

	return nil
}

func (a *App) subscribeRedirect(ctx context.Context, host, path string) (string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path,
	}

	log.Printf("subscribing to %s", u.String())

	bs, err := json.Marshal(&operation.ReqCommand{
		Command: command.Subscribe,
	})
	if err != nil {
		return "", fmt.Errorf("marshal ReqCommand failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), bytes.NewReader(bs))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout:       httpClientTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %w", err)
	}

	if cerr := resp.Body.Close(); cerr != nil {
		return "", fmt.Errorf("failed to close: %w", cerr)
	}

	loc, err := resp.Location()
	if errors.Is(err, http.ErrNoLocation) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed to get location: %w", err)
	}

	return loc.String(), nil
}

func (a *App) sendCommand(commandType command.Type) error {
	if commandType == command.Subscribe {
		return nil
	}

	log.Printf("sending %s command", commandType)

	b, err := json.Marshal(&operation.ReqCommand{
		Command: commandType,
	})
	if err != nil {
		return fmt.Errorf("marshal ReqCommand failed: %w", err)
	}

	if err := a.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return fmt.Errorf("write binary message failed: %w", err)
	}

	return nil
}

// readWs must be invoked in goroutine.
func (a *App) readWs() {
	defer close(a.done)

	for {
		_, message, err := a.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}

			return
		}

		log.Printf("recv: %s", message)
	}
}

func (a *App) writeWs() {
	ticker := time.NewTicker(sendCommandsAfter)
	defer ticker.Stop()

	for {
		select {
		case <-a.done:
			return
		case <-ticker.C:
			if err := a.sendCommand(command.NumConnections); err != nil {
				log.Printf("send command failed: %v", err)

				return
			}

			time.Sleep(timeoutBeforeUnsubscribe)

			if err := a.sendCommand(command.Unsubscribe); err != nil {
				log.Printf("send command failed: %v", err)

				return
			}

			return
		case <-a.interrupt:
			log.Println("interrupt")

			if err := a.conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				log.Printf("write close failed: %v", err)

				return
			}
			select {
			case <-a.done:
			case <-time.After(time.Second):
			}

			return
		}
	}
}
