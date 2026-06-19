package websocket

import "github.com/atendi9/ws/io/pkg/websocket/wsconn"

const TextMessage = wsconn.TextMessage
const PingMessage = wsconn.PingMessage
const PongMessage = wsconn.PongMessage
const BinaryMessage = wsconn.BinaryMessage
const CloseMessage = wsconn.CloseMessage

func FormatCloseMessage(code int, message string) []byte {
	return wsconn.FormatCloseMessage(code, message)
}

type PreparedMessage = wsconn.PreparedMessage

func NewPreparedMessage(mt int, msg []byte) (*PreparedMessage, error) {
	return wsconn.NewPreparedMessage(mt, msg)
}
