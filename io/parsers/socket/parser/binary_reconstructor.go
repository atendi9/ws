package parser

import (
	"sync"
	"sync/atomic"

	"github.com/atendi9/ws/io/pkg/forge"
)

// binaryReconstructor manages the reconstruction of binary packets.
// It collects binary buffers until the expected number of attachments
// is received, then reconstructs the complete packet.
type binaryReconstructor struct {
	mu      sync.Mutex
	buffers []forge.Interface
	packet  atomic.Pointer[Packet]
}

// NewBinaryReconstructor creates a new binaryReconstructor for the given packet.
// The packet should have its Attachments field set to the expected number of buffers.
func NewBinaryReconstructor(packet *Packet) *binaryReconstructor {
	br := &binaryReconstructor{
		buffers: make([]forge.Interface, 0),
	}
	br.packet.Store(packet)
	return br
}

// WithBinaryData adds a binary buffer to the reconstruction.
// Returns the reconstructed packet when all expected buffers are received,
// or nil if more buffers are needed.
func (br *binaryReconstructor) WithBinaryData(data forge.Interface) (*Packet, error) {
	br.mu.Lock()
	defer br.mu.Unlock()

	br.buffers = append(br.buffers, data)

	packet := br.packet.Load()
	if packet == nil || packet.Attachments == nil {
		return nil, nil
	}

	if uint64(len(br.buffers)) == *packet.Attachments {
		reconstructedPacket, err := ReconstructPacket(packet, br.buffers)
		br.Reset()
		return reconstructedPacket, err
	}

	return nil, nil
}

// Reset clears the reconstruction state.
func (br *binaryReconstructor) Reset() {
	br.buffers = nil
	br.packet.Store(nil)
}

// Finished signals that reconstruction is complete or canceled.
// This should be called when the decoder is destroyed or reset.
func (br *binaryReconstructor) Finished() {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.Reset()
}
