package websocket

type WSMessage struct {
	conn    *WSConn
	MsgType string
	Data    interface{}
}

func (msg *WSMessage) Conn() *WSConn {
	return msg.conn
}

func NewMessage(conn *WSConn, MsgType string, Data interface{}) *WSMessage {
	return &WSMessage{conn, MsgType, Data}
}
