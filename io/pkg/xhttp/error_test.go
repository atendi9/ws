package xhttp

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestCodeMessage(t *testing.T) {
	code := 404
	message := "Not Found"

	cm := CodeMessage{
		Code:    code,
		Message: message,
	}

	assert.Equal(t, code, cm.Code)
	assert.Equal(t, message, cm.Message)
}

func TestErrorMessage(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	httpContext := NewContext(w, r)

	ctxData := map[string]any{"user_id": 123}

	errMsg := ErrorMessage{
		CodeMessage: &CodeMessage{
			Code:    500,
			Message: "Internal Server Error",
		},
		Req:     httpContext,
		Context: ctxData,
	}

	assert.Equal(t, 500, errMsg.Code)
	assert.Equal(t, "Internal Server Error", errMsg.Message)
	assert.Equal(t, httpContext, errMsg.Req)
	assert.LengthMap(t, 1, errMsg.Context)
	assert.Equal(t, 123, errMsg.Context["user_id"])
}

func TestExtendedError(t *testing.T) {
	msg := "failed to process request"
	data := map[string]string{"reason": "timeout"}

	extErr := NewExtendedError(msg, data)

	// Test Error interface implementation
	assert.Equal(t, msg, extErr.Error())

	// Test Data field
	assert.Equal(t, fmt.Sprint(data), fmt.Sprint(extErr.Data))

	// Test Err method returns the error interface
	err := extErr.Err()
	assert.Error(t, err)
	assert.Equal(t, msg, err.Error())
}
