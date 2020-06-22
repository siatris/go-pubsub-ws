package websocket

import (
	"context"
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

type WSHandler func(WSMessage)

type WSMessage interface {
	Conn() WSConn
	Data() interface{}
	Type() string

	json.Marshaler
	json.Unmarshaler
}

type WSReadWriter interface {
	ReceiveMessage() (WSMessage, error)
	SendMessage(namespace string, data interface{}) error
}

type WSConn interface {
	AuthClaims() interface{}
	ID() uuid.UUID
	WriteMessage(msg WSMessage) error
	AddHandler(t string, handler WSHandler) int
	RemoveHandler(t string, id int)
	Handle(ctx context.Context)

	WSReadWriter
	io.Closer
}

type WSService interface {
	HandleConnect(func(conn WSConn))
	Run(context.Context, string)
}
