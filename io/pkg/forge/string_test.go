package forge

import (
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestString(t *testing.T) {
	t.Run("NewFromString", func(t *testing.T) {
		sb := NewFromString("hello world")
		strBuf, ok := sb.(*String)
		assert.True(t, ok)
		assert.Equal(t, "hello world", strBuf.GoString())
	})

	t.Run("NewStringReader", func(t *testing.T) {
		r := strings.NewReader("reader data")
		sb, err := NewStringReader(r)

		assert.NoError(t, err)

		strBuf, ok := sb.(*String)
		assert.True(t, ok)
		assert.Equal(t, "reader data", strBuf.GoString())
	})

	t.Run("NewString", func(t *testing.T) {
		sb := NewString([]byte("byte data"))
		strBuf, ok := sb.(*String)
		assert.True(t, ok)
		assert.Equal(t, "byte data", strBuf.GoString())
	})

	t.Run("Clone", func(t *testing.T) {
		sb := &String{NewBufferString("original")}
		clone := sb.Clone()

		strBufClone, ok := clone.(*String)
		assert.True(t, ok)
		assert.Equal(t, "original", strBufClone.GoString())
	})

	t.Run("Clone_Nil", func(t *testing.T) {
		var sb *String
		clone := sb.Clone()
		assert.Equal(t, nil, clone)
	})

	t.Run("GoString", func(t *testing.T) {
		sb := &String{NewBufferString("gostring content")}
		assert.Equal(t, "gostring content", sb.GoString())
	})

	t.Run("GoString_Nil", func(t *testing.T) {
		var sb *String
		assert.Equal(t, "<nil>", sb.GoString())
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		sb := &String{NewBufferString("json content")}
		data, err := sb.MarshalJSON()

		assert.NoError(t, err)
		assert.Equal(t, `"json content"`, string(data))
	})

	t.Run("MarshalJSON_Nil", func(t *testing.T) {
		var sb *String
		data, err := sb.MarshalJSON()

		assert.NoError(t, err)
		assert.Equal(t, `""`, string(data))
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		sb := &String{}
		err := sb.UnmarshalJSON([]byte(`"unmarshaled content"`))

		assert.NoError(t, err)
		assert.Equal(t, "unmarshaled content", sb.GoString())
	})

	t.Run("UnmarshalJSON_Error", func(t *testing.T) {
		sb := &String{}
		err := sb.UnmarshalJSON([]byte(`invalid json`))
		assert.Error(t, err)
	})

	t.Run("UnmarshalJSON_NilReceiver", func(t *testing.T) {
		var sb *String
		err := sb.UnmarshalJSON([]byte(`"test"`))
		assert.NoError(t, err)
	})
}
