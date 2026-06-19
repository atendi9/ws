package forge

import "bytes"

// IndexByte returns the index of the first instance of c in b, or -1 if c is not present in b.
// It is a wrapper around [bytes.IndexByte].
func IndexByte(b []byte, c byte) int {
	return bytes.IndexByte(b, c)
}
