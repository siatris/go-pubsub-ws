package pubsub

import (
	"context"
	"fmt"
	"go-pubsub-ws/pkg/websocket"
	"sync"

	"github.com/go-redis/redis/v8"
)

type Middleware func(*websocket.WSMessage) *websocket.WSMessage

type Subscription interface {
	Subscriber() *websocket.WSConn
	Use(Middleware)
	Namespaces() []string
	Handle(*websocket.WSMessage) *websocket.WSMessage
	ReceiveMessage(ctx context.Context) *websocket.WSMessage
}

type subscription struct {
	conn        *websocket.WSConn
	namespaces  []string
	middlewares []Middleware
	redisSub    *redis.PubSub
	sync.RWMutex
}

func (s *subscription) ReceiveMessage(ctx context.Context) *websocket.WSMessage {
	msgi, err := s.redisSub.Receive(ctx)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	switch msg := msgi.(type) {
	case *redis.Message:
		return websocket.NewMessage(s.conn, msg.Channel, msg.Payload)
	default:
		return nil
	}
}

func (s *subscription) Subscriber() *websocket.WSConn {
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

func (s *subscription) Handle(msg *websocket.WSMessage) *websocket.WSMessage {
	var m *websocket.WSMessage = msg
	for _, middleware := range s.middlewares {
		m = middleware(m)
	}
	return m
}
