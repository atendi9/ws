// Package errors defines custom error structures and predefined error instances
// used across the application components.
package errors

import (
	"context"
	"fmt"
)

// Error represents a detailed engine error containing additional context,
// a specific error type classification, and underlying wrapped errors.
type Error struct {
	// Message holds the human-readable error reason.
	Message string
	// Description wraps the underlying cause of this [Error].
	Description error
	// Type categorizes the error (e.g., "TransportError", "ParserError").
	Type string
	// Context retains the [context.Context] associated with where the error occurred.
	Context context.Context
	// errs contains the slice of internal errors for unwrapping.
	errs []error
}

// NewEngineError creates and returns a new pointer to an [Error] with the provided
// [context.Context], error classification type, message reason, and underlying description error.
func NewEngineError(
	context context.Context,
	errorType string,
	reason string,
	description error,
) *Error {
	return &Error{
		Message:     reason,
		Description: description,
		Type:        errorType,
		Context:     context,
		errs:        []error{description},
	}
}

// Err returns the [Error] instance itself as a standard error interface.
func (e *Error) Err() error {
	return e
}

// Error formats the [Error] into a string representation matching the pattern "Type::Message".
func (e *Error) Error() string {
	return fmt.Sprintf("%s::%s", e.Type, e.Message)
}

// Unwrap returns the underlying slice of errors contained within this [Error].
func (e *Error) Unwrap() []error {
	return e.errs
}

// NewTransportError is a helper that creates a specific [Error] categorized as a "TransportError".
func NewTransportError(
	ctx context.Context,
	reason string,
	description error,
) *Error {
	return NewEngineError(ctx, "TransportError", reason, description)
}

var (
	// Context defines the default base [context.Context] used for package-level predefined errors.
	Context = context.Background()

	// ErrParserPacketNil indicates that a parser encountered a nil packet.
	ErrParserPacketNil = NewEngineError(Context, "ParserError", "packet must not be nil", nil)
	// ErrParserPacketType indicates that an invalid packet type was supplied to the parser.
	ErrParserPacketType = NewEngineError(Context, "ParserError", "invalid packet type", nil)
	// ErrParserDataNil indicates that the data payload provided to the parser is nil.
	ErrParserDataNil = NewEngineError(Context, "ParserError", "data must not be nil", nil)
	// ErrParserInvalidDataLength indicates that the data length does not match expectations.
	ErrParserInvalidDataLength = NewEngineError(Context, "ParserError", "invalid data length", nil)
	// ErrParser represents a generic parsing failure.
	ErrParser = NewEngineError(Context, "ParserError", "parsing error", nil)
	// ErrParserUnknownPacketType indicates that the packet type cannot be recognized by the parser.
	ErrParserUnknownPacketType = NewEngineError(Context, "ParserError", "unknown packet type", nil)

	// ErrTryTransportAltServices indicates that all alternative service connections failed.
	ErrTryTransportAltServices = NewEngineError(Context, "TransportError", "all alt-svc attempts failed", nil)
	// ErrOriginMatch indicates that the alternative service endpoint hostname does not match the origin.
	ErrOriginMatch = NewEngineError(Context, "TransportError", "alt-svc endpoint hostname does not match origin", nil)
	// ErrEndpointInvalidCharacters indicates that an alternative service endpoint contains illegal characters.
	ErrEndpointInvalidCharacters = NewEngineError(Context, "TransportError", "alt-svc endpoint contains invalid characters", nil)
	// ErrInvalidEndpointFormat indicates an invalid format for the alternative service endpoint.
	ErrInvalidEndpointFormat = NewEngineError(Context, "TransportError", "invalid alt-svc endpoint format", nil)
	// ErrNetworkConnectionLost indicates that the underlying network connection was dropped.
	ErrNetworkConnectionLost = NewEngineError(Context, "TransportError", "network connection lost", nil)
	// ErrNoTransportsAvailable indicates that no valid transport mechanisms are available.
	ErrNoTransportsAvailable = NewEngineError(Context, "TransportError", "no transports available", nil)
	// ErrUnsupportedTransport indicates that the specified transport name is not recognized.
	ErrUnsupportedTransport = NewEngineError(Context, "TransportError", "unsupported transport name", nil)
	// ErrTransportClosed indicates that the active transport channel has been closed.
	ErrTransportClosed = NewEngineError(Context, "TransportError", "transport closed", nil)
	// ErrTransportClosedByTheServer indicates that the connection was closed from the remote server side.
	ErrTransportClosedByTheServer = NewEngineError(Context, "TransportError", "transport closed by the server", nil)
	// ErrTransportNotImplemented indicates that the requested transport type is not implemented.
	ErrTransportNotImplemented = NewEngineError(Context, "TransportError", "transport creation not implemented", nil)
	// ErrTransportFailure indicates a generic failure within the transport layer.
	ErrTransportFailure = NewEngineError(Context, "TransportError", "transport failure", nil)
	// ErrReadLimit indicates that the webtransport read threshold has been exceeded.
	ErrReadLimit = NewEngineError(Context, "TransportError", "webtransport: read limit exceeded", nil)
	// ErrBadWriteOpCode indicates that an invalid write message operation code was encountered.
	ErrBadWriteOpCode = NewEngineError(Context, "TransportError", "webtransport: bad write message type", nil)
	// ErrWriteClosed indicates that writing operations are no longer allowed because it is closed.
	ErrWriteClosed = NewEngineError(Context, "TransportError", "webtransport: write closed", nil)

	// ErrTypesBufferSeekInvalidWhence indicates an invalid whence value was passed during a seek operation.
	ErrTypesBufferSeekInvalidWhence = NewEngineError(Context, "BufferError", "types.Buffer.Seek: invalid whence", nil)
	// ErrTypesBufferSeekNegativePosition indicates an attempt to seek to a negative position index.
	ErrTypesBufferSeekNegativePosition = NewEngineError(Context, "BufferError", "types.Buffer.Seek: negative position", nil)

	// ErrHTTPContextResponseAlreadyWritten indicates an action was performed after headers/body were already sent.
	ErrHTTPContextResponseAlreadyWritten = NewEngineError(Context, "HTTPContextError", "response has already been written", nil)
	// ErrHTTPContextInvalidStatusCode indicates that an invalid HTTP response status code was supplied.
	ErrHTTPContextInvalidStatusCode = NewEngineError(Context, "HTTPContextError", "invalid status code", nil)
	// ErrHTTPContextNilRequest indicates that the standard [net/http.Request] reference is nil.
	ErrHTTPContextNilRequest = NewEngineError(Context, "HTTPContextError", "http.Request must not be nil", nil)
	// ErrHTTPContextNilResponseWriter indicates that the standard [net/http.ResponseWriter] reference is nil.
	ErrHTTPContextNilResponseWriter = NewEngineError(Context, "HTTPContextError", "http.ResponseWriter must not be nil", nil)

	// ErrSliceEmpty indicates that the target slice contains no elements.
	ErrSliceEmpty = NewEngineError(Context, "SliceError", "slice is empty", nil)
	// ErrSliceIndexOutOfBounds indicates that the requested index is outside the bounds of the slice.
	ErrSliceIndexOutOfBounds = NewEngineError(Context, "SliceError", "index out of bounds", nil)
	// ErrInvalidSliceRange indicates that the specified slice range parameters are invalid.
	ErrInvalidSliceRange = NewEngineError(Context, "SliceError", "invalid slice range", nil)

	// ErrEmptyURI indicates that the provided URI string is empty.
	ErrEmptyURI = NewEngineError(Context, "URIError", "URI must not be empty", nil)
	// ErrUnsupportedScheme indicates that the schema prefix of the URI is not supported.
	ErrUnsupportedScheme = NewEngineError(Context, "URIError", "unsupported URI scheme", nil)

	// ErrTimeout indicates a generic operation timeout constraint was hit.
	ErrTimeout = NewEngineError(Context, "TimeoutError", "timed out", nil)
	// ErrOperationHasTimedOut indicates that the ongoing execution operation exceeded its deadline.
	ErrOperationHasTimedOut = NewEngineError(Context, "TimeoutError", "operation has timed out", nil)
	// ErrUnknownServerType indicates that the specified server infrastructure configuration is unknown.
	ErrUnknownServerType = NewEngineError(Context, "ServerError", "unknown server type", nil)

	// ErrDataMustNotBeNil indicates that the handshake configuration payload data is missing.
	ErrDataMustNotBeNil = NewEngineError(Context, "HandshakeError", "data must not be nil", nil)
	// ErrHandshakeDecode indicates a failure occurred while deserializing handshake information.
	ErrHandshakeDecode = NewEngineError(Context, "HandshakeError", "failed to decode handshake data", nil)
	// ErrHandshakeSID provides compatibility warning information regarding Socket.IO v2.x and v3.x mismatch.
	ErrHandshakeSID = NewEngineError(Context, "HandshakeError", "it seems you are trying to reach a Socket.IO server in v2.x with a v3.x client, but they are not compatible (more information here: https://socket.io/docs/migrating-from-2-x-to-3-0/)", nil)

	// ErrSocketClosed indicates that the target socket session has already been closed.
	ErrSocketClosed = NewEngineError(Context, "SocketError", "socket closed", nil)
	// ErrSocketDisconnected indicates that the socket connection has been disconnected completely.
	ErrSocketDisconnected = NewEngineError(Context, "SocketError", "socket has been disconnected", nil)

	// ErrUnreadByte indicates an invalid sequence where UnreadByte was called without a preceding successful read.
	ErrUnreadByte = NewEngineError(Context, "UnreadByte", "bytes.Buffer: UnreadByte: previous operation was not a successful read", nil)
	// ErrUnreadRune indicates an invalid sequence where UnreadRune was called without a preceding successful ReadRune.
	ErrUnreadRune = NewEngineError(Context, "UnreadRune", "bytes.Buffer: UnreadRune: previous operation was not a successful ReadRune", nil)

	// ErrFetchSocketsNotSupported indicates FetchSockets operation is invoked on a parent namespace context where unsupported.
	ErrFetchSocketsNotSupported = NewEngineError(Context, "FetchSocketsError", "FetchSockets() is not supported on parent namespaces", nil)
	// ErrPayloadTooLarge indicates that the JSONP frame or request payload exceeds structural limits.
	ErrPayloadTooLarge = NewEngineError(Context, "PayloadError", "jsonp payload too large", nil)
	// ErrInvalidHeartbeat indicates an unrecognized or illegal direction state for heartbeats.
	ErrInvalidHeartbeat = NewEngineError(Context, "InvalidHeartbeat", "invalid heartbeat direction", nil)
	// ErrIllegalAttachments indicates that unexpected or prohibited attachments were found in the packet.
	ErrIllegalAttachments = NewEngineError(Context, "IllegalAttachments", "illegal attachments", nil)
	// ErrTooLarge indicates that the internal buffer allocations exceeded maximum thresholds.
	ErrTooLarge = NewEngineError(Context, "BufferTooLarge", "bytes.Buffer: too large", nil)
	// ErrNegativeRead indicates that the source reader illegally returned a negative count number.
	ErrNegativeRead = NewEngineError(Context, "NegativeRead", "bytes.Buffer: reader returned negative count from Read", nil)
	// ErrCloseSent indicates that a close request notification has already been transmitted over WebTransport.
	ErrCloseSent = NewEngineError(Context, "CloseSent", "webtransport: close sent", nil)
	// ErrPlaintextDuringReconstruction indicates unencrypted data was received during partial packet stream reassembly.
	ErrPlaintextDuringReconstruction = NewEngineError(Context, "PlaintextDuringReconstruction", "got plaintext data when reconstructing a packet", nil)
	// ErrBinaryWithoutReconstruction indicates binary fragments were received outside of an active packet assembly context.
	ErrBinaryWithoutReconstruction = NewEngineError(Context, "BinaryWithoutReconstruction", "got binary data when not reconstructing a packet", nil)
	// ErrInvalidPayload indicates that the transmission frame layout or content payload format is invalid.
	ErrInvalidPayload = NewEngineError(Context, "InvalidPayload", "invalid payload", nil)
	// ErrIllegalNamespace indicates access to an invalid or unauthenticated namespace identification route.
	ErrIllegalNamespace = NewEngineError(Context, "IllegalNamespace", "illegal namespace", nil)
	// ErrIllegalID indicates that the message or session identification sequence is format-broken or malicious.
	ErrIllegalID = NewEngineError(Context, "IllegalID", "illegal id", nil)
	// ErrTooManyAttachments indicates that the message packet exceeds the allowed structural attachments quota.
	ErrTooManyAttachments = NewEngineError(Context, "TooManyAttachments", "too many attachments", nil)
)
