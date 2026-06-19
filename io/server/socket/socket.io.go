package socket

import (
	"github.com/atendi9/ws/io/server/engine/transports"
)

type (
	WebSocketBuilder = transports.WebSocketBuilder
	PollingBuilder   = transports.PollingBuilder
	TConstructor     = transports.TConstructor
)

var (
	Polling   TConstructor = &PollingBuilder{}
	WebSocket TConstructor = &WebSocketBuilder{}
)
