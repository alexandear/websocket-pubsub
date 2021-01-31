// Code generated by MockGen. DO NOT EDIT.
// Source: client.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	server "github.com/alexandear/websocket-pubsub/internal/server"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockHubI is a mock of HubI interface
type MockHubI struct {
	ctrl     *gomock.Controller
	recorder *MockHubIMockRecorder
}

// MockHubIMockRecorder is the mock recorder for MockHubI
type MockHubIMockRecorder struct {
	mock *MockHubI
}

// NewMockHubI creates a new mock instance
func NewMockHubI(ctrl *gomock.Controller) *MockHubI {
	mock := &MockHubI{ctrl: ctrl}
	mock.recorder = &MockHubIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHubI) EXPECT() *MockHubIMockRecorder {
	return m.recorder
}

// Subscribe mocks base method
func (m *MockHubI) Subscribe(client server.ClientI) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Subscribe", client)
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockHubIMockRecorder) Subscribe(client interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockHubI)(nil).Subscribe), client)
}

// Unsubscribe mocks base method
func (m *MockHubI) Unsubscribe(client server.ClientI) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unsubscribe", client)
}

// Unsubscribe indicates an expected call of Unsubscribe
func (mr *MockHubIMockRecorder) Unsubscribe(client interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockHubI)(nil).Unsubscribe), client)
}

// Cast mocks base method
func (m *MockHubI) Cast(data server.CastData) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Cast", data)
}

// Cast indicates an expected call of Cast
func (mr *MockHubIMockRecorder) Cast(data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Cast", reflect.TypeOf((*MockHubI)(nil).Cast), data)
}

// Run mocks base method
func (m *MockHubI) Run(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Run", ctx)
}

// Run indicates an expected call of Run
func (mr *MockHubIMockRecorder) Run(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockHubI)(nil).Run), ctx)
}

// MockWsConn is a mock of WsConn interface
type MockWsConn struct {
	ctrl     *gomock.Controller
	recorder *MockWsConnMockRecorder
}

// MockWsConnMockRecorder is the mock recorder for MockWsConn
type MockWsConnMockRecorder struct {
	mock *MockWsConn
}

// NewMockWsConn creates a new mock instance
func NewMockWsConn(ctrl *gomock.Controller) *MockWsConn {
	mock := &MockWsConn{ctrl: ctrl}
	mock.recorder = &MockWsConnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWsConn) EXPECT() *MockWsConnMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockWsConn) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockWsConnMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockWsConn)(nil).Close))
}

// ReadBinaryMessage mocks base method
func (m *MockWsConn) ReadBinaryMessage() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadBinaryMessage")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadBinaryMessage indicates an expected call of ReadBinaryMessage
func (mr *MockWsConnMockRecorder) ReadBinaryMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadBinaryMessage", reflect.TypeOf((*MockWsConn)(nil).ReadBinaryMessage))
}

// WriteBinaryMessage mocks base method
func (m *MockWsConn) WriteBinaryMessage(data []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteBinaryMessage", data)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteBinaryMessage indicates an expected call of WriteBinaryMessage
func (mr *MockWsConnMockRecorder) WriteBinaryMessage(data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteBinaryMessage", reflect.TypeOf((*MockWsConn)(nil).WriteBinaryMessage), data)
}

// WriteCloseMessage mocks base method
func (m *MockWsConn) WriteCloseMessage() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WriteCloseMessage")
}

// WriteCloseMessage indicates an expected call of WriteCloseMessage
func (mr *MockWsConnMockRecorder) WriteCloseMessage() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteCloseMessage", reflect.TypeOf((*MockWsConn)(nil).WriteCloseMessage))
}
