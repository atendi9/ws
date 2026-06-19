package anvil

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNew(t *testing.T) {
	ptr := New("")
	assert.NotNil(t, ptr)
	assert.Equal(t, "", *ptr)
	msg := "Hello World"
	ptr = New(msg)
	assert.NotNil(t, ptr)
	assert.Equal(t, msg, *ptr)

	var user struct {
		Name string
	}
	usrPtr := New(user)
	assert.NotNil(t, usrPtr)
	assert.Equal(t, "", user.Name)

	usrName := "John Doe"
	usrPtr = New(struct{ Name string }{Name: usrName})
	assert.NotNil(t, usrPtr)
	assert.Equal(t, usrName, usrPtr.Name)
}
