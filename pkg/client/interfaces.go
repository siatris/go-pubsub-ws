package client

import (
	"io"

	"github.com/siatris/go-pubsub-ws/pkg/pubsub"
	"github.com/siatris/go-pubsub-ws/pkg/websocket"
)

type Client interface {
	io.Closer
	websocket.WSReadWriter
	pubsub.PubSubber
}
