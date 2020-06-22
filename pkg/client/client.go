package client

import (
	"context"
	"fmt"
	"net/url"

	"go-pubsub-ws/pkg/pubsub"
	"go-pubsub-ws/pkg/websocket"

	gowebsocket "github.com/gorilla/websocket"
)

type WithOption func(*client)

type client struct {
	token string
	host  string
	ctx   context.Context

	websocket.WSConn
}

func Connect(opts ...WithOption) (Client, error) {
	cli := &client{}

	for _, opt := range opts {
		opt(cli)
	}

	u := url.URL{Scheme: "ws", Host: cli.host, Path: "/"}
	query := u.Query()
	query.Set("token", cli.token)
	u.RawQuery = query.Encode()

	c, _, err := gowebsocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	if cli.ctx == nil {
		cli.ctx = context.Background()
	}

	cli.WSConn = websocket.WrapConnection(c)

	go cli.WSConn.Handle(cli.ctx)

	return cli, nil
}

func (c *client) Close() error {
	return c.WSConn.Close()
}

func (c *client) Subscribe(namespace string, handler websocket.WSHandler) (func(), error) {
	subErrReceiver := make(chan error, 0)
	subType := fmt.Sprintf(pubsub.SUBSCRIBED_TYPE, namespace)

	subHandler := c.AddHandler(subType, func(msg websocket.WSMessage) {
		subErrReceiver <- nil
	})

	unsubscribe := func() {}

	defer func() {
		c.RemoveHandler(subType, subHandler)
	}()

	err := c.SendMessage(pubsub.SUBSCRIBE_TYPE, map[string]string{
		"Namespace": namespace,
	})
	if err != nil {
		return unsubscribe, err
	}

	if subErr := <-subErrReceiver; subErr != nil {
		return unsubscribe, subErr
	}

	handleID := c.AddHandler(namespace, handler)
	unsubscribe = func() {
		c.RemoveHandler(namespace, handleID)
	}

	return unsubscribe, nil
}

func (c *client) Publish(namespace string, data interface{}) error {
	return c.SendMessage(pubsub.PUBLISH_TYPE, map[string]interface{}{
		"Namespace": namespace,
		"Data":      data,
	})
}

func WithHost(host string) WithOption {
	return func(c *client) {
		c.host = host
	}
}

func WithAuthToken(token string) WithOption {
	return func(c *client) {
		c.token = token
	}
}

func WithContext(ctx context.Context) WithOption {
	return func(c *client) {
		c.ctx = ctx
	}
}
