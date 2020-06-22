package pubsub

import (
	"go-pubsub-ws/pkg/websocket"

	"github.com/go-redis/redis/v8"
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
