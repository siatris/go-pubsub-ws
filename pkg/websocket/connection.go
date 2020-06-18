package websocket

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WSTypeHandler func(*WSMessage)

type WSConn struct {
	conn       *websocket.Conn
	uid        uuid.UUID
	handlers   map[string][]WSTypeHandler
	authClaims interface{}
	sync.RWMutex
}

func (c *WSConn) AuthClaims() interface{} {
	return c.authClaims
}

func (c *WSConn) ID() uuid.UUID {
	return c.uid
}

func (c *WSConn) WriteMessage(msg *WSMessage) error {
	return c.conn.WriteJSON([]interface{}{msg.MsgType, msg.Data})
}

func (c *WSConn) AddHandler(t string, handler WSTypeHandler) int {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	if _, ok := c.handlers[t]; !ok {
		c.handlers[t] = make([]WSTypeHandler, 0)
	}
	c.handlers[t] = append(c.handlers[t], handler)
	return len(c.handlers[t]) - 1
}

func (c *WSConn) RemoveHandler(t string, id int) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	if _, ok := c.handlers[t]; !ok {
		return
	}
	c.handlers[t] = append(c.handlers[t][:id], c.handlers[t][id+1:]...)
}

func (c *WSConn) Handle(ctx context.Context) {
	defer c.conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg []interface{}
			err := c.conn.ReadJSON(&msg)
			if err != nil {
				return
			}
			if len(msg) != 2 {
				continue
			}
			t := msg[0].(string)
			c.RWMutex.RLock()
			if handlers, ok := c.handlers[t]; ok {
				for _, h := range handlers {
					var handler func(*WSMessage) = h
					go handler(&WSMessage{
						conn:    c,
						MsgType: t,
						Data:    msg[1],
					})
				}
			}
			c.RWMutex.RUnlock()
		}
	}
}
