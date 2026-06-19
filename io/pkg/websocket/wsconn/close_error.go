package wsconn

import (
	"io"
	"strconv"
)

const (
	// CloseNormalClosure indicates a normal closure, meaning that the purpose for
	// which the connection was established has been fulfilled.
	CloseNormalClosure = 1000
	// CloseGoingAway indicates that an endpoint is "going away", such as a server
	// going down or a browser having navigated away from a page.
	CloseGoingAway = 1001
	// CloseProtocolError indicates that an endpoint is terminating the connection
	// due to a protocol error.
	CloseProtocolError = 1002
	// CloseUnsupportedData indicates that an endpoint is terminating the connection
	// because it has received a type of data it cannot accept.
	CloseUnsupportedData = 1003
	// CloseNoStatusReceived is a reserved value and MUST NOT be set as a status code in a
	// Close control frame by an endpoint. It indicates no status code was provided.
	CloseNoStatusReceived = 1005
	// CloseAbnormalClosure is a reserved value and MUST NOT be set as a status code in a
	// Close control frame by an endpoint. It indicates the connection was closed abnormally.
	CloseAbnormalClosure = 1006
	// CloseInvalidFramePayloadData indicates that an endpoint is terminating the connection
	// because it has received data within a message that was not consistent with the type of the message.
	CloseInvalidFramePayloadData = 1007
	// ClosePolicyViolation indicates that an endpoint is terminating the connection
	// because it has received a message that violates its policy.
	ClosePolicyViolation = 1008
	// CloseMessageTooBig indicates that an endpoint is terminating the connection
	// because it has received a message that is too big for it to process.
	CloseMessageTooBig = 1009
	// CloseMandatoryExtension indicates that an endpoint (client) is terminating the
	// connection because it has expected the server to negotiate one or more extensions.
	CloseMandatoryExtension = 1010
	// CloseInternalServerErr indicates that a server is terminating the connection because
	// it encountered an unexpected condition that prevented it from fulfilling the request.
	CloseInternalServerErr = 1011
	// CloseServiceRestart indicates that the service is restarting. A client may reconnect,
	// and if it chooses to do so, should reconnect using a randomized delay.
	CloseServiceRestart = 1012
	// CloseTryAgainLater indicates that the service is experiencing overload. A client should
	// only connect to a different IP (when there are multiple for the target) or reconnect
	// to the same IP upon user action.
	CloseTryAgainLater = 1013
	// CloseTLSHandshake is a reserved value and MUST NOT be set as a status code in a
	// Close control frame by an endpoint. It indicates that the connection was closed
	// due to a failure to perform a TLS handshake.
	CloseTLSHandshake = 1015
)

// CloseError represents an error containing a WebSocket close code and a
// corresponding descriptive text message.
type CloseError struct {
	// Code is the integer status code defined in RFC 6455.
	Code int
	// Text is the optional message associated with the connection closure.
	Text string
}

// Error implements the error interface, returning a formatted string
// that includes the closure code and its standard textual representation.
func (e *CloseError) Error() string {
	s := []byte("websocket: close ")
	s = strconv.AppendInt(s, int64(e.Code), 10)
	switch e.Code {
	case CloseNormalClosure:
		s = append(s, "<normal>"...)
	case CloseGoingAway:
		s = append(s, "<going away>"...)
	case CloseProtocolError:
		s = append(s, "<protocol error>"...)
	case CloseUnsupportedData:
		s = append(s, "<unsupported data>"...)
	case CloseNoStatusReceived:
		s = append(s, "<no status>"...)
	case CloseAbnormalClosure:
		s = append(s, "<abnormal closure>"...)
	case CloseInvalidFramePayloadData:
		s = append(s, "<invalid payload data>"...)
	case ClosePolicyViolation:
		s = append(s, "<policy violation>"...)
	case CloseMessageTooBig:
		s = append(s, "<message too big>"...)
	case CloseMandatoryExtension:
		s = append(s, "<mandatory extension missing>"...)
	case CloseInternalServerErr:
		s = append(s, "<internal server error>"...)
	case CloseTLSHandshake:
		s = append(s, "<TLS handshake error>"...)
	}
	if e.Text != "" {
		s = append(s, ": "...)
		s = append(s, e.Text...)
	}
	return string(s)
}

// ErrUnexpectedEOF represents a generic abnormal closure error, typically
// returned when the underlying connection experiences an unexpected EOF.
var ErrUnexpectedEOF = &CloseError{Code: CloseAbnormalClosure, Text: io.ErrUnexpectedEOF.Error()}

// IsUnexpectedCloseError returns true if the provided error is of type *CloseError.
// It can be used to determine if the connection was closed with a standard WebSocket error.
func IsUnexpectedCloseError(err error) bool {
	if _, ok := err.(*CloseError); ok {
		return true
	}
	return false
}
