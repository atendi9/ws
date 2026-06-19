package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/atendi9/ws/io/pkg/anvil"
	"github.com/atendi9/ws/io/pkg/errors"
	"github.com/atendi9/ws/io/pkg/events"
	"github.com/atendi9/ws/io/pkg/forge"
)

const (
	// DefaultMaxAttachments is the default maximum number of binary attachments allowed per [Packet].
	// This prevents resource exhaustion from malicious clients sending excessively large attachment counts.
	DefaultMaxAttachments uint64 = 10

	// DefaultMaxNamespaceLength is the default maximum allowed length of a namespace name.
	// This prevents resource exhaustion from malicious clients sending excessively long namespace names.
	DefaultMaxNamespaceLength int = 512

	// DefaultMaxPacketIDLength is the default maximum allowed length of a [Packet] ID string.
	// uint64 max is 18446744073709551615 (20 digits).
	DefaultMaxPacketIDLength int = 20
)

// ReservedEvents contains event names that have special meaning in Socket.IO
// and cannot be used as custom event names.
var ReservedEvents = anvil.NewSet(
	"connect",        // Used on the client side to indicate connection
	"connect_error",  // Used on the client side to indicate connection error
	"disconnect",     // Used on both sides to indicate disconnection
	"disconnecting",  // Used on the server side during disconnection
	"newListener",    // Used by the Node.js Emitter
	"removeListener", // Used by the Node.js Emitter
)

// decoder implements the [Decoder] interface for Socket.IO packet decoding.
type decoder struct {
	events.Emitter

	// reconstructor manages binary [Packet] reconstruction state.
	reconstructor atomic.Pointer[binaryReconstructor]

	opts DecoderOptions
}

// NewDecoder creates a new [Decoder] instance.
// An optional [DecoderOptions] can be provided to configure the [decoder].
func NewDecoder(opts ...DecoderOptions) Decoder {
	options := DefaultDecoderOptions()
	options.SetMaxAttachments(DefaultMaxAttachments)
	options.SetMaxNamespaceLength(DefaultMaxNamespaceLength)
	options.SetMaxPacketIDLength(DefaultMaxPacketIDLength)

	if len(opts) > 0 && opts[0] != nil {
		options.Assign(opts[0])
	}

	return &decoder{
		Emitter: events.NewEmitter(),
		opts:    options,
	}
}

// Add processes incoming data (string or binary) and emits decoded packets.
// For string data, it decodes immediately. For binary data, it accumulates
// buffers until the [Packet] is complete, then emits the reconstructed [Packet].
func (d *decoder) Add(data any) error {
	switch typedData := data.(type) {
	case string:
		return d.processTextBuffer(forge.NewFromString(typedData))

	case *strings.Reader:
		buffer, err := forge.NewStringReader(typedData)
		if err != nil {
			return err
		}
		return d.processTextBuffer(buffer)

	case *forge.String:
		return d.processTextBuffer(typedData)

	default:
		return d.processBinaryBuffer(data)
	}
}

// processTextBuffer processes string-based [Packet] data via an [forge.Interface].
func (d *decoder) processTextBuffer(buffer forge.Interface) error {
	if d.reconstructor.Load() != nil {
		return errors.ErrPlaintextDuringReconstruction
	}
	return d.decodeTextBuffer(buffer)
}

// processBinaryBuffer processes binary [Packet] data for reconstruction.
func (d *decoder) processBinaryBuffer(data any) error {
	if !IsBinary(data) {
		return fmt.Errorf("unknown type: %T", data)
	}

	reconstructor := d.reconstructor.Load()
	if reconstructor == nil {
		return errors.ErrBinaryWithoutReconstruction
	}

	buffer, err := d.readIntoBuffer(data)
	if err != nil {
		return err
	}

	packet, err := reconstructor.WithBinaryData(buffer)
	if err != nil {
		return fmt.Errorf("decode error: %w", err)
	}

	if packet != nil {
		// Received final buffer, [Packet] is complete
		d.reconstructor.Store(nil)
		d.Emit("decoded", packet)
	}

	return nil
}

// readIntoBuffer reads binary data from various source types into an [forge.Interface].
func (d *decoder) readIntoBuffer(data any) (forge.Interface, error) {
	buffer := forge.NewBytesBuffer(nil)

	switch typedData := data.(type) {
	case io.Reader:
		if closer, ok := data.(io.Closer); ok {
			defer func() {
				if err := closer.Close(); err != nil {
				}
			}()
		}
		if _, err := buffer.ReadFrom(typedData); err != nil {
			return nil, err
		}
	case []byte:
		if _, err := buffer.Write(typedData); err != nil {
			return nil, err
		}
	}

	return buffer, nil
}

// decodeTextBuffer decodes a string [forge.Interface] and handles binary [Packet] initialization.
func (d *decoder) decodeTextBuffer(buffer forge.Interface) error {
	packet, err := d.parsePacket(buffer)
	if err != nil {
		return err
	}

	if packet.Type == BINARY_EVENT || packet.Type == BINARY_ACK {
		d.reconstructor.Store(NewBinaryReconstructor(packet))
		// If no attachments expected, emit immediately
		if packet.Attachments != nil && *packet.Attachments == 0 {
			d.Emit("decoded", packet)
		}
	} else {
		// Non-binary [Packet], emit immediately
		d.Emit("decoded", packet)
	}

	return nil
}

// parsePacket parses a [Packet] from a string [forge.Interface].
func (d *decoder) parsePacket(buffer forge.Interface) (*Packet, error) {
	packet := &Packet{}

	// Parse packet type
	if err := d.readPacketType(buffer, packet); err != nil {
		return nil, err
	}

	// Parse attachments for binary packets
	if err := d.readAttachmentCount(buffer, packet); err != nil {
		return nil, err
	}

	// Parse namespace
	if err := d.readNamespace(buffer, packet); err != nil {
		return nil, err
	}

	// Parse packet ID
	if err := d.readPacketID(buffer, packet); err != nil {
		return nil, err
	}

	// Parse payload data
	if err := d.readPayload(buffer, packet); err != nil {
		return nil, err
	}
	return packet, nil
}

// readPacketType reads and validates the [PacketType] for a [Packet].
func (d *decoder) readPacketType(buffer forge.Interface, packet *Packet) error {
	typeByte, err := buffer.ReadByte()
	if err != nil {
		return errors.ErrInvalidPayload
	}

	packet.Type = PacketType(int(typeByte) - '0')
	if !packet.Type.Valid() {
		return fmt.Errorf("unknown packet type %d", packet.Type)
	}

	return nil
}

// readAttachmentCount reads the attachment count for binary [Packet] instances.
func (d *decoder) readAttachmentCount(buffer forge.Interface, packet *Packet) error {
	if packet.Type != BINARY_EVENT && packet.Type != BINARY_ACK {
		return nil
	}

	attachmentStr, err := buffer.ReadString('-')
	if err != nil {
		return errors.ErrIllegalAttachments
	}

	strLen := len(attachmentStr)
	if strLen < 2 { // Must be at least "X-" where X is a digit
		return errors.ErrIllegalAttachments
	}

	attachmentCount, err := strconv.ParseUint(attachmentStr[:strLen-1], 10, 64)
	if err != nil {
		return errors.ErrIllegalAttachments
	}

	if attachmentCount > d.opts.MaxAttachments() {
		return errors.ErrTooManyAttachments
	}

	packet.Attachments = &attachmentCount
	return nil
}

// readNamespace reads the namespace from the [forge.Interface] into the [Packet].
func (d *decoder) readNamespace(buffer forge.Interface, packet *Packet) error {
	firstByte, err := buffer.ReadByte()
	if err != nil {
		if err == io.EOF {
			packet.Nsp = "/"
			return nil
		}
		return errors.ErrIllegalNamespace
	}

	if firstByte != '/' {
		// No namespace specified, use default and put byte back
		if unreadErr := buffer.UnreadByte(); unreadErr != nil {
			return errors.ErrIllegalNamespace
		}
		packet.Nsp = "/"
		return nil
	}

	// Read the rest of the namespace until comma
	nspSuffix, err := buffer.ReadString(',')
	if err != nil {
		if err == io.EOF {
			if len(nspSuffix)+1 > d.opts.MaxNamespaceLength() {
				return errors.ErrIllegalNamespace
			}
			packet.Nsp = "/" + nspSuffix
			return nil
		}
		return errors.ErrIllegalNamespace
	}

	// Remove trailing comma
	nsp := "/" + nspSuffix[:len(nspSuffix)-1]
	if len(nsp) > d.opts.MaxNamespaceLength() {
		return errors.ErrIllegalNamespace
	}
	packet.Nsp = nsp
	return nil
}

// readPacketID reads the optional packet ID for acknowledgments into the [Packet].
func (d *decoder) readPacketID(buffer forge.Interface, packet *Packet) error {
	if buffer.Len() == 0 {
		return nil
	}

	var idBuilder strings.Builder

	for {
		if idBuilder.Len() >= d.opts.MaxPacketIDLength() {
			return errors.ErrIllegalID
		}

		b, err := buffer.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if b >= '0' && b <= '9' {
			if err := idBuilder.WriteByte(b); err != nil {
				return err
			}
		} else {
			if err := buffer.UnreadByte(); err != nil {
				return errors.ErrIllegalID
			}
			break
		}
	}

	if idBuilder.Len() > 0 {
		packetID, err := strconv.ParseUint(idBuilder.String(), 10, 64)
		if err != nil {
			return err
		}
		packet.Id = &packetID
	}

	return nil
}

// readPayload reads and validates the JSON payload into the [Packet].
func (d *decoder) readPayload(buffer forge.Interface, packet *Packet) error {
	if buffer.Len() == 0 {
		return d.validateData(packet.Type, nil)
	}

	var payload any
	if err := json.NewDecoder(buffer).Decode(&payload); err != nil {
		return errors.ErrInvalidPayload
	}

	if err := d.validateData(packet.Type, payload); err != nil {
		return err
	}

	packet.Data = payload
	return nil
}

// validateData checks if the payload is valid for the given [PacketType].
func (d *decoder) validateData(packetType PacketType, payload any) error {
	if !isDataValid(packetType, payload) {
		return errors.ErrInvalidPayload
	}
	return nil
}

// Destroy releases the [decoder] resources and stops any ongoing reconstruction.
func (d *decoder) Destroy() {
	if reconstructor := d.reconstructor.Load(); reconstructor != nil {
		reconstructor.Finished()
	}
}

// isDataValid checks if the payload matches the expected format for the [PacketType].
func isDataValid(packetType PacketType, payload any) bool {
	switch packetType {
	case CONNECT:
		return payload == nil || isMap(payload)
	case DISCONNECT:
		return payload == nil
	case CONNECT_ERROR:
		return isMap(payload) || isString(payload)
	case EVENT, BINARY_EVENT:
		return isValidEventPayload(payload)
	case ACK, BINARY_ACK:
		return isSlice(payload)
	default:
		return false
	}
}

// isMap checks if the payload is a map[string]any.
func isMap(payload any) bool {
	_, ok := payload.(map[string]any)
	return ok
}

// isString checks if the payload is a string.
func isString(payload any) bool {
	_, ok := payload.(string)
	return ok
}

// isSlice checks if the payload is a slice.
func isSlice(payload any) bool {
	_, ok := payload.([]any)
	return ok
}

// isValidEventPayload validates that an event payload has a valid event name.
// The event name can be either a string (not in reserved events) or a number.
func isValidEventPayload(payload any) bool {
	data, ok := payload.([]any)
	if !ok || len(data) == 0 {
		return false
	}

	// Event name can be a string or a number
	switch Name := data[0].(type) {
	case string:
		return !ReservedEvents.Has(Name)
	case float64: // JSON numbers are decoded as float64 in Go
		return true
	case int, int64, int32:
		return true
	default:
		return false
	}
}
