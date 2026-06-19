package parser

import (
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/ws/io/pkg/forge"
)

func TestPacketParser(t *testing.T) {
	t.Run("DeconstructPacket", func(t *testing.T) {
		packet := &Packet{
			Data: []any{
				"regular string",
				[]byte("binary data"),
			},
		}

		resultPacket, buffers := DeconstructPacket(packet)
		assert.LengthSlice(t, 1, buffers)

		assert.True(t, resultPacket.Attachments != nil)
		assert.Equal(t, uint64(1), *resultPacket.Attachments)

		sliceData, ok := resultPacket.Data.([]any)
		assert.True(t, ok)
		assert.LengthSlice(t, 2, sliceData)

		placeholder, ok := sliceData[1].(*Placeholder)
		assert.True(t, ok)
		assert.True(t, placeholder.Placeholder)
		assert.Equal(t, int64(0), placeholder.Num)
	})

	t.Run("ReconstructPacket", func(t *testing.T) {
		packet := &Packet{
			Data: map[string]any{
				"file": map[string]any{
					"_placeholder": true,
					"num":          float64(0),
				},
			},
			Attachments: new(uint64),
		}

		buf := forge.NewBytesBuffer(nil)
		_, _ = buf.Write([]byte("reconstructed data"))
		buffers := []forge.Interface{buf}

		resultPacket, err := ReconstructPacket(packet, buffers)
		assert.NoError(t, err)

		assert.True(t, resultPacket.Attachments == nil)

		mapData, ok := resultPacket.Data.(map[string]any)
		assert.True(t, ok)
		assert.LengthMap(t, 1, mapData)

		extractedBuf, ok := mapData["file"].(forge.Interface)
		assert.True(t, ok)
		assert.Equal(t, buf, extractedBuf)
	})

	t.Run("ReconstructPacket_ErrorOutofBounds", func(t *testing.T) {
		packet := &Packet{
			Data: map[string]any{
				"file": map[string]any{
					"_placeholder": true,
					"num":          float64(99),
				},
			},
		}

		buf := forge.NewBytesBuffer(nil)
		buffers := []forge.Interface{buf}

		_, err := ReconstructPacket(packet, buffers)
		assert.Error(t, err)
	})
}
