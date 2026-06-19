package xrequest

import (
	"io"
)

type dummyReader struct {
	body io.ReadCloser
}

// Read reads bytes from the underlying stream.
func (d *dummyReader) Read(p []byte) (n int, err error) {
	return d.body.Read(p)
}

// Close closes the underlying stream.
func (d *dummyReader) Close() error {
	return d.body.Close()
}
