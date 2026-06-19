package errors

import (
	"context"
	"errors"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestErrors(t *testing.T) {
	t.Run("NewEngineError structural validation", func(t *testing.T) {
		ctx := context.Background()
		innerErr := errors.New("low level failure")

		result := NewEngineError(ctx, "TestEngine", "failed to process engine", innerErr)

		assert.Equal(t, "failed to process engine", result.Message)
		assert.Equal(t, "TestEngine", result.Type)
		assert.Equal(t, innerErr, result.Description)
		assert.Equal(t, ctx, result.Context)
	})

	t.Run("Error implementation and methods", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		innerErr := errors.New("underlying issue")

		errInstance := NewEngineError(ctx, "CustomType", "main message", innerErr)

		assert.Equal(t, "CustomType::main message", errInstance.Error())

		unwrapped := errInstance.Unwrap()
		assert.LengthSlice(t, 1, unwrapped)
		assert.Equal(t, innerErr, unwrapped[0])
	})

	t.Run("NewTransportError formatting", func(t *testing.T) {
		ctx := context.Background()
		innerErr := errors.New("connection timeout")

		result := NewTransportError(ctx, "mapping failure", innerErr)

		assert.Equal(t, "TransportError", result.Type)
		assert.Equal(t, "mapping failure", result.Message)
		assert.Equal(t, innerErr, result.Description)
		assert.Equal(t, ctx, result.Context)
	})

	t.Run("Predefined Parser errors", func(t *testing.T) {
		assert.Equal(t, "ParserError::packet must not be nil", ErrParserPacketNil.Error())
		assert.Equal(t, "ParserError::invalid packet type", ErrParserPacketType.Error())
		assert.Equal(t, "ParserError::data must not be nil", ErrParserDataNil.Error())
		assert.Equal(t, "ParserError::invalid data length", ErrParserInvalidDataLength.Error())
		assert.Equal(t, "ParserError::parsing error", ErrParser.Error())
		assert.Equal(t, "ParserError::unknown packet type", ErrParserUnknownPacketType.Error())
	})

	t.Run("Predefined Transport errors", func(t *testing.T) {
		assert.Equal(t, "TransportError::all alt-svc attempts failed", ErrTryTransportAltServices.Error())
		assert.Equal(t, "TransportError::alt-svc endpoint hostname does not match origin", ErrOriginMatch.Error())
		assert.Equal(t, "TransportError::alt-svc endpoint contains invalid characters", ErrEndpointInvalidCharacters.Error())
		assert.Equal(t, "TransportError::invalid alt-svc endpoint format", ErrInvalidEndpointFormat.Error())
		assert.Equal(t, "TransportError::network connection lost", ErrNetworkConnectionLost.Error())
		assert.Equal(t, "TransportError::no transports available", ErrNoTransportsAvailable.Error())
		assert.Equal(t, "TransportError::unsupported transport name", ErrUnsupportedTransport.Error())
		assert.Equal(t, "TransportError::transport closed", ErrTransportClosed.Error())
		assert.Equal(t, "TransportError::transport closed by the server", ErrTransportClosedByTheServer.Error())
		assert.Equal(t, "TransportError::transport creation not implemented", ErrTransportNotImplemented.Error())
		assert.Equal(t, "TransportError::transport failure", ErrTransportFailure.Error())
		assert.Equal(t, "TransportError::webtransport: read limit exceeded", ErrReadLimit.Error())
		assert.Equal(t, "TransportError::webtransport: bad write message type", ErrBadWriteOpCode.Error())
		assert.Equal(t, "TransportError::webtransport: write closed", ErrWriteClosed.Error())
	})

	t.Run("Predefined Buffer and HTTPContext errors", func(t *testing.T) {
		assert.Equal(t, "BufferError::types.Buffer.Seek: invalid whence", ErrTypesBufferSeekInvalidWhence.Error())
		assert.Equal(t, "BufferError::types.Buffer.Seek: negative position", ErrTypesBufferSeekNegativePosition.Error())

		assert.Equal(t, "HTTPContextError::response has already been written", ErrHTTPContextResponseAlreadyWritten.Error())
		assert.Equal(t, "HTTPContextError::invalid status code", ErrHTTPContextInvalidStatusCode.Error())
		assert.Equal(t, "HTTPContextError::http.Request must not be nil", ErrHTTPContextNilRequest.Error())
		assert.Equal(t, "HTTPContextError::http.ResponseWriter must not be nil", ErrHTTPContextNilResponseWriter.Error())
	})

	t.Run("Predefined Slice, URI, Timeout and Server errors", func(t *testing.T) {
		assert.Equal(t, "SliceError::slice is empty", ErrSliceEmpty.Error())
		assert.Equal(t, "SliceError::index out of bounds", ErrSliceIndexOutOfBounds.Error())
		assert.Equal(t, "SliceError::invalid slice range", ErrInvalidSliceRange.Error())

		assert.Equal(t, "URIError::URI must not be empty", ErrEmptyURI.Error())
		assert.Equal(t, "URIError::unsupported URI scheme", ErrUnsupportedScheme.Error())

		assert.Equal(t, "TimeoutError::timed out", ErrTimeout.Error())
		assert.Equal(t, "TimeoutError::operation has timed out", ErrOperationHasTimedOut.Error())
		assert.Equal(t, "ServerError::unknown server type", ErrUnknownServerType.Error())
	})

	t.Run("Predefined Handshake and Socket errors", func(t *testing.T) {
		assert.Equal(t, "HandshakeError::data must not be nil", ErrDataMustNotBeNil.Error())
		assert.Equal(t, "HandshakeError::failed to decode handshake data", ErrHandshakeDecode.Error())
		assert.True(t, len(ErrHandshakeSID.Error()) > 0)

		assert.Equal(t, "SocketError::socket closed", ErrSocketClosed.Error())
		assert.Equal(t, "SocketError::socket has been disconnected", ErrSocketDisconnected.Error())
	})

	t.Run("Predefined Miscellaneous internal errors", func(t *testing.T) {
		assert.Equal(t, "UnreadByte::bytes.Buffer: UnreadByte: previous operation was not a successful read", ErrUnreadByte.Error())
		assert.Equal(t, "UnreadRune::bytes.Buffer: UnreadRune: previous operation was not a successful ReadRune", ErrUnreadRune.Error())
		assert.Equal(t, "FetchSocketsError::FetchSockets() is not supported on parent namespaces", ErrFetchSocketsNotSupported.Error())
		assert.Equal(t, "PayloadError::jsonp payload too large", ErrPayloadTooLarge.Error())
		assert.Equal(t, "InvalidHeartbeat::invalid heartbeat direction", ErrInvalidHeartbeat.Error())
		assert.Equal(t, "IllegalAttachments::illegal attachments", ErrIllegalAttachments.Error())
		assert.Equal(t, "BufferTooLarge::bytes.Buffer: too large", ErrTooLarge.Error())
		assert.Equal(t, "NegativeRead::bytes.Buffer: reader returned negative count from Read", ErrNegativeRead.Error())
		assert.Equal(t, "CloseSent::webtransport: close sent", ErrCloseSent.Error())
		assert.Equal(t, "PlaintextDuringReconstruction::got plaintext data when reconstructing a packet", ErrPlaintextDuringReconstruction.Error())
		assert.Equal(t, "BinaryWithoutReconstruction::got binary data when not reconstructing a packet", ErrBinaryWithoutReconstruction.Error())
		assert.Equal(t, "InvalidPayload::invalid payload", ErrInvalidPayload.Error())
		assert.Equal(t, "IllegalNamespace::illegal namespace", ErrIllegalNamespace.Error())
		assert.Equal(t, "IllegalID::illegal id", ErrIllegalID.Error())
		assert.Equal(t, "TooManyAttachments::too many attachments", ErrTooManyAttachments.Error())
	})
}
