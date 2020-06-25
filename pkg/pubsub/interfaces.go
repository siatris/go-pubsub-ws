package pubsub

import (
	"github.com/go-redis/redis/v8"
	"github.com/siatris/go-pubsub-ws/pkg/websocket"
)

type RedisProvider interface {
	redis.UniversalClient
}

type Publisher interface {
	Publish(namespace string, data interface{}) error
}

type Subscriber interface {
	Subscribe(namespace string, handler websocket.WSHandler) (func(), error)
}

type PubSubber interface {
	Publisher
	Subscriber
}
