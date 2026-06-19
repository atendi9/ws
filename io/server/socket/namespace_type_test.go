package socket

import (
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestNewExtendedError(t *testing.T) {
	err := NewExtendedError("nope", map[string]any{"code": 1})
	assert.NotNil(t, err)
	assert.Equal(t, "nope", err.Error())
	assert.Error(t, err.Err())
}

func TestSessionData_GetPid(t *testing.T) {
	tests := []struct {
		name    string
		data    *SessionData
		wantPid string
		wantOk  bool
	}{
		{"string pid", &SessionData{Pid: "abc"}, "abc", true},
		{"slice pid takes last", &SessionData{Pid: []string{"a", "b", "z"}}, "z", true},
		{"empty slice", &SessionData{Pid: []string{}}, "", false},
		{"nil pid", &SessionData{Pid: nil}, "", false},
		{"empty string", &SessionData{Pid: ""}, "", false},
		{"unsupported type", &SessionData{Pid: 123}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pid, ok := tt.data.GetPid()
			assert.Equal(t, tt.wantPid, pid)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestSessionData_GetPidNilReceiver(t *testing.T) {
	var s *SessionData
	pid, ok := s.GetPid()
	assert.Equal(t, "", pid)
	assert.False(t, ok)
}

func TestSessionData_GetOffset(t *testing.T) {
	tests := []struct {
		name       string
		data       *SessionData
		wantOffset string
		wantOk     bool
	}{
		{"string offset", &SessionData{Offset: "off1"}, "off1", true},
		{"slice offset takes last", &SessionData{Offset: []string{"x", "y"}}, "y", true},
		{"empty slice", &SessionData{Offset: []string{}}, "", false},
		{"nil offset", &SessionData{Offset: nil}, "", false},
		{"unsupported type", &SessionData{Offset: 3.14}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			off, ok := tt.data.GetOffset()
			assert.Equal(t, tt.wantOffset, off)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestSessionData_GetOffsetNilReceiver(t *testing.T) {
	var s *SessionData
	off, ok := s.GetOffset()
	assert.Equal(t, "", off)
	assert.False(t, ok)
}
