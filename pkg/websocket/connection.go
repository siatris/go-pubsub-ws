package websocket

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func WrapConnection(conn *websocket.Conn) *wsconn {
	return &wsconn{
		conn:     conn,
		uid:      uuid.New(),
		handlers: make(map[string][]WSHandler),
	}
}

type wsconn struct {
	conn       *websocket.Conn
	uid        uuid.UUID
	handlers   map[string][]WSHandler
	authClaims interface{}
	sync.RWMutex
}

func (c *wsconn) AuthClaims() interface{} {
	return c.authClaims
}

func (c *wsconn) ID() uuid.UUID {
	return c.uid
}

func (c *wsconn) WriteMessage(msg WSMessage) error {
	return c.conn.WriteJSON(msg)
}

func (c *wsconn) AddHandler(t string, handler WSHandler) int {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	if _, ok := c.handlers[t]; !ok {
		c.handlers[t] = make([]WSHandler, 0)
	}
	c.handlers[t] = append(c.handlers[t], handler)
	return len(c.handlers[t]) - 1
}

func (c *wsconn) RemoveHandler(t string, id int) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	if _, ok := c.handlers[t]; !ok {
		return
	}
	c.handlers[t] = append(c.handlers[t][:id], c.handlers[t][id+1:]...)
}

func (c *wsconn) Handle(ctx context.Context) {
	defer c.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg wsmessage
			err := c.conn.ReadJSON(&msg)
			if err != nil {
				return
			}
			t := msg.msgType
			c.RWMutex.RLock()
			if handlers, ok := c.handlers[t]; ok {
				for _, h := range handlers {
					var handler func(WSMessage) = h
					go handler(&wsmessage{
						conn:    c,
						msgType: t,
						data:    msg.data,
					})
				}
			}

			if wildcardHandlers, ok := c.handlers["*"]; ok {
				for _, h := range wildcardHandlers {
					var handler func(WSMessage) = h
					go handler(&wsmessage{
						conn:    c,
						msgType: t,
						data:    msg.data,
					})
				}
			}

			c.RWMutex.RUnlock()
		}
	}
}

func (c *wsconn) ReceiveMessage() (WSMessage, error) {
	var msg wsmessage

	err := c.conn.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (c *wsconn) SendMessage(mtype string, data interface{}) error {
	return c.conn.WriteJSON(NewMessage(c, mtype, data))
}

func (c *wsconn) Close() error {
	return c.conn.Close()
}
