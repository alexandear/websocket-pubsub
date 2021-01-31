package client_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/alexandear/websocket-pubsub/internal/client"
	"github.com/alexandear/websocket-pubsub/internal/client/mock"
	"github.com/alexandear/websocket-pubsub/internal/pkg/operation"
	"github.com/alexandear/websocket-pubsub/internal/pkg/websocket"
)

func TestClient_ReadOne(t *testing.T) {
	t.Run("when does not set conn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()

		resp, err := cl.ReadOne()

		assert.Nil(t, resp)
		assert.EqualError(t, err, client.ErrNilConn.Error())
	})

	t.Run("when closed conn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()
		connm := mock.NewMockWsConn(ctrl)
		cl.SetConn(connm)
		connm.EXPECT().ReadBinaryMessage().Return(nil, websocket.ErrClosedConn).Times(1)

		resp, err := cl.ReadOne()

		assert.Nil(t, resp)
		assert.EqualError(t, err, websocket.ErrClosedConn.Error())
	})

	t.Run("when broadcast", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()
		connm := mock.NewMockWsConn(ctrl)
		cl.SetConn(connm)
		id := uuid.New().String()
		ts := int(time.Now().Unix())
		connm.EXPECT().ReadBinaryMessage().Return([]byte(
			fmt.Sprintf(`{"client_id":"%s","timestamp":%d}`, id, ts)), nil).Times(1)

		resp, err := cl.ReadOne()

		assert.NoError(t, err)
		assert.Equal(t, operation.RespBroadcast{
			ClientID:  id,
			Timestamp: ts,
		}, resp)
	})

	t.Run("when num connections", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()
		connm := mock.NewMockWsConn(ctrl)
		cl.SetConn(connm)
		numConns := rand.Intn(100) + 1
		connm.EXPECT().ReadBinaryMessage().Return([]byte(
			fmt.Sprintf(`{"num_connections":%d}`, numConns)), nil).Times(1)

		resp, err := cl.ReadOne()

		assert.NoError(t, err)
		assert.Equal(t, operation.RespNumConnections{
			NumConnections: numConns,
		}, resp)
	})
}

func TestClient_Subscribe(t *testing.T) {
	t.Run("when does not set conn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()

		err := cl.Subscribe()

		assert.EqualError(t, err, client.ErrNilConn.Error())
	})

	t.Run("when ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()
		connm := mock.NewMockWsConn(ctrl)
		connm.EXPECT().WriteBinaryMessage([]byte(`{"command":"SUBSCRIBE"}`)).Times(1)

		cl.SetConn(connm)
		err := cl.Subscribe()

		assert.NoError(t, err)
	})
}

func TestClient_NumConnections(t *testing.T) {
	t.Run("when does not set conn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()

		err := cl.NumConnections()

		assert.EqualError(t, err, client.ErrNilConn.Error())
	})

	t.Run("when ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()
		connm := mock.NewMockWsConn(ctrl)
		connm.EXPECT().WriteBinaryMessage([]byte(`{"command":"NUM_CONNECTIONS"}`)).Times(1)

		cl.SetConn(connm)
		err := cl.NumConnections()

		assert.NoError(t, err)
	})
}

func TestClient_Unsubscribe(t *testing.T) {
	t.Run("when does not set conn", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()

		err := cl.Unsubscribe()

		assert.EqualError(t, err, client.ErrNilConn.Error())
	})

	t.Run("when ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cl := client.NewClient()
		connm := mock.NewMockWsConn(ctrl)
		connm.EXPECT().WriteBinaryMessage([]byte(`{"command":"UNSUBSCRIBE"}`)).Times(1)

		cl.SetConn(connm)
		err := cl.Unsubscribe()

		assert.NoError(t, err)
	})
}
