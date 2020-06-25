package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/siatris/go-pubsub-ws/pkg/websocket"

	"github.com/go-redis/redis/v8"
)

type Middleware func(websocket.WSMessage) websocket.WSMessage

type Subscription interface {
	Subscriber() websocket.WSConn
	Use(Middleware)
	Namespaces() []string
	Handle(websocket.WSMessage) websocket.WSMessage
	ReceiveMessage(ctx context.Context) (websocket.WSMessage, error)
}

type subscription struct {
	conn        websocket.WSConn
	namespaces  []string
	middlewares []Middleware
	sub         *redis.PubSub
	sync.RWMutex
}

func (s *subscription) ReceiveMessage(ctx context.Context) (websocket.WSMessage, error) {
	msgi, err := s.sub.Receive(ctx)
	if err != nil {
		return nil, err
	}
	switch msg := msgi.(type) {
	case *redis.Message:
		return websocket.NewMessage(s.conn, msg.Channel, msg.Payload), nil
	case *redis.Subscription:
		return websocket.NewMessage(s.conn, fmt.Sprintf(SUBSCRIBED_TYPE, msg.Channel), msg), nil
	default:
		return nil, errors.New("Invalid message")
	}
}

func (s *subscription) Subscriber() websocket.WSConn {
	return s.conn
}

func (s *subscription) Namespaces() []string {
	s.RLock()
	defer s.RUnlock()
	return s.namespaces
}

func (s *subscription) Use(middleware Middleware) {
	s.Lock()
	defer s.Unlock()
	s.middlewares = append(s.middlewares, middleware)
}

func (s *subscription) Handle(msg websocket.WSMessage) websocket.WSMessage {
	var m websocket.WSMessage = msg
	for _, middleware := range s.middlewares {
		m = middleware(m)
	}
	return m
}
