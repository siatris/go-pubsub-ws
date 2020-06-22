package client

import (
	"io"

	"go-pubsub-ws/pkg/pubsub"
	"go-pubsub-ws/pkg/websocket"
)

type Client interface {
	io.Closer
	websocket.WSReadWriter
	pubsub.PubSubber
}
