package pubsub

import (
	"context"
	"sync"

	"github.com/siatris/go-pubsub-ws/pkg/websocket"
)

type PubSubService struct {
	wss         websocket.WSService
	rdb         RedisProvider
	subHandlers map[string]func(conn websocket.WSConn) (Subscription, error)
	pubHandlers map[string]func(msg *Publication) error
	sync.RWMutex
}

func New(opts ...WithOption) *PubSubService {
	pss := &PubSubService{
		subHandlers: make(map[string]func(conn websocket.WSConn) (Subscription, error), 0),
		pubHandlers: make(map[string]func(msg *Publication) error, 0),
	}

	for _, opt := range opts {
		opt(pss)
	}

	return pss
}

type WithOption func(*PubSubService)

func WithRedisProvider(redisProvider RedisProvider) WithOption {
	return func(s *PubSubService) {
		s.rdb = redisProvider
	}
}

func WithWSService(ws websocket.WSService) WithOption {
	return func(s *PubSubService) {
		s.wss = ws
	}
}

func (ps *PubSubService) HandleSubscription(namespace string, handler func(conn websocket.WSConn) (Subscription, error)) {
	ps.subHandlers[namespace] = handler
}

func (ps *PubSubService) HandlePublication(namespace string, handler func(msg *Publication) error) {
	ps.pubHandlers[namespace] = handler
}

func (ps *PubSubService) Run(ctx context.Context) {
	if ps.wss != nil {
		ps.wss.HandleConnect(func(conn websocket.WSConn) {
			conn.AddHandler(SUBSCRIBE_TYPE, func(msg websocket.WSMessage) {
				subMsg := msg.Data().(map[string]interface{})
				var sub Subscription

				for namespace, handler := range ps.subHandlers {
					if namespace == subMsg["Namespace"].(string) {
						var err error
						sub, err = handler(msg.Conn())
						if err != nil {
							return
						}
						break
					}
				}

				if sub == nil {
					return
				}

				for {
					msg, err := sub.ReceiveMessage(ctx)
					if err != nil {
						continue
					}

					if msg != nil {
						msg = sub.Handle(msg)
						conn.WriteMessage(msg)
					}

					select {
					case <-ctx.Done():
						return
					default:
					}
				}
			})

			conn.AddHandler(PUBLISH_TYPE, func(msg websocket.WSMessage) {
				pub := msg.Data().(map[string]interface{})
				ps.PublishTo(ctx, msg.Conn(), pub["Namespace"].(string), pub["Data"])
			})

			conn.Handle(ctx)
		})

		ps.wss.Run(ctx)
	}
}

func (ps *PubSubService) CreateSubscription(ctx context.Context, conn websocket.WSConn, namespaces ...string) Subscription {
	redisSub := ps.rdb.Subscribe(ctx, namespaces...)
	return &subscription{conn: conn, namespaces: namespaces, sub: redisSub, middlewares: make([]Middleware, 0)}
}

func (ps *PubSubService) PublishTo(ctx context.Context, origin websocket.WSConn, namespace string, data interface{}) error {
	pub := &Publication{
		origin: origin,
		data:   data,
	}

	for namespace, handler := range ps.pubHandlers {
		if namespace == namespace {
			err := handler(pub)
			if err != nil {
				return err
			}
		}
	}

	ps.rdb.Publish(ctx, namespace, data)
	return nil
}
