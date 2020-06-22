package websocket

import (
	"encoding/json"
)

type wsmessage struct {
	conn    WSConn
	msgType string
	data    interface{}
}

func (msg *wsmessage) Conn() WSConn {
	return msg.conn
}

func (msg *wsmessage) Type() string {
	return msg.msgType
}

func (msg *wsmessage) Data() interface{} {
	return msg.data
}

func (msg *wsmessage) UnmarshalJSON(b []byte) error {
	var arr []interface{}
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	msg.msgType = arr[0].(string)
	msg.data = arr[1]
	return nil
}

func (msg *wsmessage) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{msg.msgType, msg.data})
}

func NewMessage(conn WSConn, msgType string, data interface{}) WSMessage {
	return &wsmessage{conn, msgType, data}
}
