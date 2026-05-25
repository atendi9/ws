package io

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/atendi9/capivara/assert"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestParamsDecoder_Success(t *testing.T) {
	payload := []byte(`{"name": "Gabriel", "age": 28}`)

	decoderFunc := func(data []byte) (User, error) {
		var u User
		err := json.Unmarshal(data, &u)
		return u, err
	}

	expected := User{
		Name: "Gabriel",
		Age:  28,
	}

	result, err := ParamsDecoder(payload, decoderFunc)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestParamsDecoder_Error(t *testing.T) {
	payload := []byte(`{"name": "Invalid"}`)
	expectedErr := errors.New("mock decoding error")

	decoderFunc := func(data []byte) (User, error) {
		return User{}, expectedErr
	}

	expected := User{}

	result, err := ParamsDecoder(payload, decoderFunc)

	assert.Error(t, err)
	assert.Equal(t, expected, result)
}
