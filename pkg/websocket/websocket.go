package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type wsService struct {
	incoming     chan WSConn
	connHandlers []func(WSConn)
	authFunc     func(token string, conn WSConn) (interface{}, error)
	sync.RWMutex
}

func (ws *wsService) handleNewConnection(ctx context.Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		conn := WrapConnection(c)
		if ws.authFunc != nil {
			q := r.URL.Query()
			token := q.Get("token")
			if token != "" {
				claims, err := ws.authFunc(token, conn)
				if err != nil {
					c.WriteJSON([]string{"AUTH", "FAILED"})
					c.Close()
					return
				}
				conn.authClaims = claims
			}
		}
		ws.incoming <- conn
	}
}

func (ws *wsService) HandleConnect(handle func(conn WSConn)) {
	ws.RWMutex.Lock()
	defer ws.RWMutex.Unlock()
	ws.connHandlers = append(ws.connHandlers, handle)
}

func (ws *wsService) Run(ctx context.Context, listen string) {
	http.HandleFunc("/", ws.handleNewConnection(ctx))
	go http.ListenAndServe(listen, nil)
	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-ws.incoming:
			ws.RWMutex.RLock()
			for _, handler := range ws.connHandlers {
				go handler(conn)
			}
			ws.RWMutex.RUnlock()
		}
	}
}

type WithOption func(*wsService)

func New(opts ...WithOption) WSService {
	ws := &wsService{
		connHandlers: make([]func(WSConn), 0),
		incoming:     make(chan WSConn),
	}

	for _, opt := range opts {
		opt(ws)
	}

	return ws
}

func WithAuthentication(authFunc func(token string, conn WSConn) (interface{}, error)) WithOption {
	return func(s *wsService) {
		s.authFunc = authFunc
	}
}
