package server_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/alexandear/websocket-pubsub/internal/pkg/websocket"
	"github.com/alexandear/websocket-pubsub/internal/server"
	"github.com/alexandear/websocket-pubsub/internal/server/mock"
)

func TestClient_Run(t *testing.T) {
	t.Run("read", func(t *testing.T) {
		t.Run("when closed websocket connection", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			hubm := mock.NewMockHubI(ctrl)
			connm := mock.NewMockWsConn(ctrl)
			client := server.NewClient(hubm, connm)

			hubm.EXPECT().Unsubscribe(gomock.Any()).Times(1)

			connm.EXPECT().ReadBinaryMessage().Return(nil, websocket.ErrClosedConn).Times(1)
			connm.EXPECT().Close().Times(1)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			client.Run(ctx)
			cancel()
		})

		t.Run("when commands", func(t *testing.T) {
			for name, tc := range map[string]struct {
				command      string
				hubmExpectFn func(mock *mock.MockHubI, clientID string)
			}{
				"subscribe": {
					command: "SUBSCRIBE",
					hubmExpectFn: func(mock *mock.MockHubI, clientID string) {
						mock.EXPECT().Subscribe(gomock.Any())
					},
				},
				"unsubscribe": {
					command: "UNSUBSCRIBE",
					hubmExpectFn: func(mock *mock.MockHubI, clientID string) {
						mock.EXPECT().Unsubscribe(gomock.Any())
					},
				},
				"num_connections": {
					command: "NUM_CONNECTIONS",
					hubmExpectFn: func(mock *mock.MockHubI, clientID string) {
						mock.EXPECT().Cast(server.UnicastData{ClientID: clientID})
					},
				},
			} {
				t.Run(name, func(t *testing.T) {
					ctrl := gomock.NewController(t)
					defer ctrl.Finish()
					hubm := mock.NewMockHubI(ctrl)
					connm := mock.NewMockWsConn(ctrl)
					client := server.NewClient(hubm, connm)

					tc.hubmExpectFn(hubm, client.ID())
					hubm.EXPECT().Unsubscribe(gomock.Any()).Times(1)

					connm.EXPECT().ReadBinaryMessage().Return([]byte(
						fmt.Sprintf(`{"command":"%s"}`, tc.command)), nil).Times(1)
					connm.EXPECT().ReadBinaryMessage().Return(nil, websocket.ErrClosedConn).Times(1)
					connm.EXPECT().Close().Times(1)

					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					client.Run(ctx)
					cancel()
				})
			}
		})
	})

	t.Run("write", func(t *testing.T) {
		id := uuid.New().String()
		now := time.Now()
		numConns := rand.Intn(10)
		for name, tc := range map[string]struct {
			responseMessage server.ResponseMessage
			expectedResp    string
		}{
			"when response broadcast": {
				responseMessage: server.ResponseBroadcast{ClientID: id, Time: now},
				expectedResp:    fmt.Sprintf(`{"client_id":"%s","timestamp":%d}`, id, now.Unix()),
			},
			"when response num connections": {
				responseMessage: server.ResponseUnicast{NumConnections: numConns},
				expectedResp:    fmt.Sprintf(`{"num_connections":%d}`, numConns),
			},
		} {
			t.Run(name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				hubm := mock.NewMockHubI(ctrl)
				connm := mock.NewMockWsConn(ctrl)
				client := server.NewClient(hubm, connm)

				hubm.EXPECT().Unsubscribe(gomock.Any()).Times(1)

				connm.EXPECT().ReadBinaryMessage().Return(nil, websocket.ErrClosedConn).Times(1)
				connm.EXPECT().WriteBinaryMessage([]byte(tc.expectedResp))
				connm.EXPECT().Close().Times(1)

				client.Response(tc.responseMessage)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				client.Run(ctx)
				cancel()
			})
		}
	})
}
