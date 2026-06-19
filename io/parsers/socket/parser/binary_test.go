package parser

import (
	"testing"

	"github.com/atendi9/ws/io/pkg/forge"

	"github.com/atendi9/capivara/assert"
)

func TestDeconstructPacketTopLevelBinary(t *testing.T) {
	pkt := &Packet{Type: BINARY_EVENT, Nsp: "/", Data: []byte("rawbytes")}
	out, buffers := DeconstructPacket(pkt)

	assert.Equal(t, 1, len(buffers))
	assert.NotNil(t, out.Attachments)
	assert.Equal(t, uint64(1), *out.Attachments)

	// The data is replaced by a placeholder.
	ph, ok := out.Data.(*Placeholder)
	assert.True(t, ok)
	assert.True(t, ph.Placeholder)
	assert.Equal(t, int64(0), ph.Num)
}

func TestDeconstructPacketNestedBinary(t *testing.T) {
	pkt := &Packet{
		Type: BINARY_EVENT,
		Nsp:  "/",
		Data: []any{
			"event",
			map[string]any{"file": []byte("abc"), "meta": "x"},
			[]any{[]byte("def")},
		},
	}
	_, buffers := DeconstructPacket(pkt)
	assert.Equal(t, 2, len(buffers)) // one in map, one in nested slice
}

func TestReconstructPacketRoundTrip(t *testing.T) {
	// After transport, placeholders arrive as decoded JSON maps (not the
	// *Placeholder struct produced by DeconstructPacket), so build the data in
	// that post-decode shape.
	data := []any{
		"event",
		map[string]any{"file": map[string]any{"_placeholder": true, "num": float64(0)}},
	}
	pkt := &Packet{Type: BINARY_EVENT, Nsp: "/", Data: data}
	buffers := []forge.Interface{forge.NewBytesBuffer([]byte("payload"))}

	reconstructed, err := ReconstructPacket(pkt, buffers)
	assert.NoError(t, err)
	assert.True(t, reconstructed.Attachments == nil)

	// The placeholder is now replaced by a forge.Interface buffer.
	got, ok := reconstructed.Data.([]any)
	assert.True(t, ok)
	m, ok := got[1].(map[string]any)
	assert.True(t, ok)
	_, ok = m["file"].(forge.Interface)
	assert.True(t, ok)
}

func TestReconstructPacketIllegalAttachment(t *testing.T) {
	// Placeholder referencing a buffer index that does not exist.
	pkt := &Packet{
		Type: BINARY_EVENT,
		Nsp:  "/",
		Data: map[string]any{"_placeholder": true, "num": float64(5)},
	}
	_, err := ReconstructPacket(pkt, []forge.Interface{})
	assert.Error(t, err)
}

func TestReconstructPacketNilData(t *testing.T) {
	pkt := &Packet{Type: EVENT, Nsp: "/", Data: nil}
	out, err := ReconstructPacket(pkt, nil)
	assert.NoError(t, err)
	assert.True(t, out.Data == nil)
}

func TestEncoderBinaryPath(t *testing.T) {
	enc := NewEncoder()

	// An EVENT carrying binary data is promoted to BINARY_EVENT and split
	// into a header buffer plus one binary buffer.
	pkt := &Packet{Type: EVENT, Nsp: "/", Data: []any{"hello", []byte("bin")}}
	buffers := enc.Encode(pkt)
	assert.Equal(t, 2, len(buffers))
	// Header begins with the BINARY_EVENT type digit.
	header := buffers[0].String()
	assert.Equal(t, byte(BINARY_EVENT)+'0', header[0])
}

func TestEncoderStringPath(t *testing.T) {
	enc := NewEncoder()
	pkt := &Packet{Type: EVENT, Nsp: "/", Data: []any{"hello", "world"}}
	buffers := enc.Encode(pkt)
	assert.Equal(t, 1, len(buffers)) // no binary -> single string buffer
}

func TestIsDataValid(t *testing.T) {
	tests := []struct {
		name    string
		typ     PacketType
		payload any
		want    bool
	}{
		{"connect nil", CONNECT, nil, true},
		{"connect map", CONNECT, map[string]any{"a": 1}, true},
		{"connect slice invalid", CONNECT, []any{1}, false},
		{"disconnect nil", DISCONNECT, nil, true},
		{"disconnect non-nil", DISCONNECT, "x", false},
		{"connect_error map", CONNECT_ERROR, map[string]any{}, true},
		{"connect_error string", CONNECT_ERROR, "boom", true},
		{"connect_error number invalid", CONNECT_ERROR, float64(1), false},
		{"event valid", EVENT, []any{"name", 1}, true},
		{"event number name valid", EVENT, []any{float64(3)}, true},
		{"event empty invalid", EVENT, []any{}, false},
		{"event non-slice invalid", EVENT, "x", false},
		{"ack slice", ACK, []any{1, 2}, true},
		{"ack non-slice invalid", ACK, map[string]any{}, false},
		{"binary_ack slice", BINARY_ACK, []any{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDataValid(tt.typ, tt.payload))
		})
	}
}

func TestIsMapStringSlice(t *testing.T) {
	assert.True(t, isMap(map[string]any{}))
	assert.False(t, isMap("x"))
	assert.True(t, isString("x"))
	assert.False(t, isString(1))
	assert.True(t, isSlice([]any{}))
	assert.False(t, isSlice("x"))
}

func TestIsValidEventPayload(t *testing.T) {
	assert.True(t, isValidEventPayload([]any{"custom", 1}))
	assert.True(t, isValidEventPayload([]any{float64(7)}))
	assert.False(t, isValidEventPayload([]any{}))
	assert.False(t, isValidEventPayload("not-a-slice"))
}
