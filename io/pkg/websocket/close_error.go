package websocket

import "github.com/atendi9/ws/io/pkg/websocket/wsconn"

// WebSocket close codes as defined in RFC 6455. These are re-exported from the
// underlying [wsconn] package.
const (
	CloseNormalClosure           = wsconn.CloseNormalClosure
	CloseGoingAway               = wsconn.CloseGoingAway
	CloseProtocolError           = wsconn.CloseProtocolError
	CloseUnsupportedData         = wsconn.CloseUnsupportedData
	CloseNoStatusReceived        = wsconn.CloseNoStatusReceived
	CloseAbnormalClosure         = wsconn.CloseAbnormalClosure
	CloseInvalidFramePayloadData = wsconn.CloseInvalidFramePayloadData
	ClosePolicyViolation         = wsconn.ClosePolicyViolation
	CloseMessageTooBig           = wsconn.CloseMessageTooBig
	CloseMandatoryExtension      = wsconn.CloseMandatoryExtension
	CloseInternalServerErr       = wsconn.CloseInternalServerErr
	CloseServiceRestart          = wsconn.CloseServiceRestart
	CloseTryAgainLater           = wsconn.CloseTryAgainLater
	CloseTLSHandshake            = wsconn.CloseTLSHandshake
)

// CloseError represents an error containing a WebSocket close code and a
// corresponding descriptive text message.
type CloseError = wsconn.CloseError

// ErrUnexpectedEOF represents a generic abnormal closure error, typically
// returned when the underlying connection experiences an unexpected EOF.
var ErrUnexpectedEOF = wsconn.ErrUnexpectedEOF

// IsUnexpectedCloseError returns true if the provided error is of type *CloseError.
// It can be used to determine if the connection was closed with a standard WebSocket error.
func IsUnexpectedCloseError(err error) bool {
	return wsconn.IsUnexpectedCloseError(err)
}
