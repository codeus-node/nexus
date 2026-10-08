package nexus

import (
	"net/http"

	"github.com/codeus-node/fail"
	di "github.com/codeus-node/generic-di"
)

type Websocket interface {
	GetRoute() string
	Handshake(w http.ResponseWriter, r *http.Request) *HttpError
	ClientConnect(client WebsocketClient) fail.CustomError
	CheckOrigin(r *http.Request) bool
	ReadBufferSize() int
	WriteBufferSize() int
	EnableCompression() bool
	Handle(data []byte, client WebsocketClient) fail.CustomError
}

type WebsocketClient interface {
	GetId() string
	Send(data []byte) fail.CustomError
	SendJson(data any) fail.CustomError
	Close() fail.CustomError
	Closed() *fail.ErrorList
	OnClose(toDo func() fail.CustomError)
}

type BaseWebsocket struct {
	Pool ClientPool
}

func NewBaseWebsocket() *BaseWebsocket {
	return &BaseWebsocket{
		Pool: di.Inject[ClientPool](),
	}
}

func (socket *BaseWebsocket) GetRoute() string {
	return "/ws"
}

func (socket *BaseWebsocket) Handshake(w http.ResponseWriter, r *http.Request) *HttpError {
	return nil
}

func (socket *BaseWebsocket) ClientConnect(client WebsocketClient) fail.CustomError {
	socket.Pool.Register(client)
	return nil
}

func (socket *BaseWebsocket) CheckOrigin(r *http.Request) bool {
	return true
}

func (socket *BaseWebsocket) ReadBufferSize() int {
	return 1024
}

func (socket *BaseWebsocket) WriteBufferSize() int {
	return 1024
}

func (socket *BaseWebsocket) EnableCompression() bool {
	return false
}

func (socket *BaseWebsocket) Handle(data []byte, client WebsocketClient) fail.CustomError {
	return nil
}
