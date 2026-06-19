package parser

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNewParser(t *testing.T) {
	p := NewParser()

	ok := p != nil
	assert.True(t, ok)
}

func TestParser_NewEncoder(t *testing.T) {
	p := NewParser()
	encoder := p.NewEncoder()

	ok := encoder != nil
	assert.True(t, ok)
}

func TestParser_NewDecoder(t *testing.T) {
	p := NewParser()
	decoder := p.NewDecoder()

	ok := decoder != nil
	assert.True(t, ok)
}
