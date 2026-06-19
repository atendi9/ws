package parser

import (
	"sync"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/pkg/forge"
)

func TestNewBinaryReconstructor(t *testing.T) {
	attachments := uint64(2)
	packet := &Packet{Attachments: &attachments}

	br := NewBinaryReconstructor(packet)

	assert.Equal(t, packet, br.packet.Load())
	assert.LengthSlice(t, 0, br.buffers)
}

func TestWithBinaryData_Incomplete(t *testing.T) {
	attachments := uint64(3)
	packet := &Packet{Attachments: &attachments}
	br := NewBinaryReconstructor(packet)

	res, err := br.WithBinaryData(forge.NewBytesBuffer([]byte{'H', 'E', 'L', 'L', 'O'}))

	assert.NoError(t, err)
	assert.True(t, res == nil)
	assert.LengthSlice(t, 1, br.buffers)
}

func TestWithBinaryData_Complete(t *testing.T) {
	attachments := uint64(2)
	packet := &Packet{Attachments: &attachments}
	br := NewBinaryReconstructor(packet)

	_, err := br.WithBinaryData(nil)
	assert.NoError(t, err)

	_, err = br.WithBinaryData(nil)
	assert.NoError(t, err)

	assert.LengthSlice(t, 0, br.buffers)

	var expectedPacket *Packet = nil
	assert.Equal(t, expectedPacket, br.packet.Load())
}

func TestBinaryReconstructor_Finished(t *testing.T) {
	attachments := uint64(2)
	packet := &Packet{Attachments: &attachments}
	br := NewBinaryReconstructor(packet)

	_, err := br.WithBinaryData(forge.NewBytesBuffer([]byte{'H', 'E', 'L', 'L', 'O'}))
	assert.NoError(t, err)
	assert.LengthSlice(t, 1, br.buffers)

	br.Finished()

	assert.LengthSlice(t, 0, br.buffers)

	var expectedPacket *Packet = nil
	assert.Equal(t, expectedPacket, br.packet.Load())
}

func TestBinaryReconstructor_Concurrency(t *testing.T) {
	attachments := uint64(100)
	packet := &Packet{Attachments: &attachments}
	br := NewBinaryReconstructor(packet)

	var wg sync.WaitGroup

	for range 50 {
		wg.Go(func() {
			_, err := br.WithBinaryData(forge.NewBytesBuffer([]byte{'H', 'E', 'L', 'L', 'O'}))
			assert.NoError(t, err)
		})
	}

	wg.Wait()
	assert.LengthSlice(t, 50, br.buffers)
}
