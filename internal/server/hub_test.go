package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"

	"github.com/alexandear/websocket-pubsub/internal/server"
	"github.com/alexandear/websocket-pubsub/internal/server/mock"
)

func TestHub_Run(t *testing.T) {
	t.Run("unicast", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := server.NewHub(100 * time.Second)
		clientm := mock.NewMockClientI(ctrl)
		id := uuid.New().String()
		clientm.EXPECT().ID().Return(id).Times(1)
		clientm.EXPECT().Response(server.ResponseUnicast{NumConnections: 1}).Times(1)

		go func() {
			h.Subscribe(clientm)
			h.Cast(server.UnicastData{ClientID: id})
		}()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		h.Run(ctx)
		cancel()
	})

	t.Run("broadcast", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h := server.NewHub(100 * time.Millisecond)
		clientm := mock.NewMockClientI(ctrl)
		id := uuid.New().String()
		clientm.EXPECT().ID().Return(id).AnyTimes()
		clientm.EXPECT().Response(gomock.Any()).AnyTimes()

		go func() {
			h.Subscribe(clientm)
		}()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		h.Run(ctx)
		cancel()
	})
}
