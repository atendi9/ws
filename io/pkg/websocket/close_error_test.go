// Package websocket provides constants and error types for WebSocket connections.
package websocket

import (
	"errors"
	"io"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestCloseError(t *testing.T) {

	t.Run("should format error correctly with known code and text", func(t *testing.T) {
		err := &CloseError{
			Code: CloseNormalClosure,
			Text: "connection closed normally",
		}
		expected := "websocket: close 1000<normal>: connection closed normally"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("should format error correctly with known code but no text", func(t *testing.T) {
		err := &CloseError{
			Code: CloseGoingAway,
		}
		expected := "websocket: close 1001<going away>"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("should format error correctly with an unknown code", func(t *testing.T) {
		err := &CloseError{
			Code: 9999,
			Text: "some unknown reason",
		}
		expected := "websocket: close 9999: some unknown reason"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("should return true when IsUnexpectedCloseError receives a CloseError", func(t *testing.T) {
		err := &CloseError{
			Code: CloseAbnormalClosure,
			Text: "abnormal behavior",
		}
		ok := IsUnexpectedCloseError(err)
		assert.True(t, ok)
	})

	t.Run("should return false when IsUnexpectedCloseError receives a generic error", func(t *testing.T) {
		err := errors.New("a generic error occurred")
		ok := IsUnexpectedCloseError(err)
		assert.False(t, ok)
	})

	t.Run("should validate the properties of the global ErrUnexpectedEOF", func(t *testing.T) {
		ok := IsUnexpectedCloseError(ErrUnexpectedEOF)
		assert.True(t, ok)

		// Type assertion to validate internal fields
		closeErr, typeOk := any(ErrUnexpectedEOF).(*CloseError)
		assert.True(t, typeOk)
		assert.Equal(t, CloseAbnormalClosure, closeErr.Code)
		assert.Equal(t, io.ErrUnexpectedEOF.Error(), closeErr.Text)
	})
}
