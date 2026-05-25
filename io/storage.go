package io

import "fmt"

const (
	// B represents a single byte.
	B StorageSize = 1
)

const (
	_ = iota // Ignore the first value (0)

	// KB represents a kilobyte (1024 bytes).
	KB StorageSize = 1 << (10 * iota)

	// MB represents a megabyte (1024 kilobytes).
	MB

	// GB represents a gigabyte (1024 megabytes).
	GB

	// TB represents a terabyte (1024 gigabytes).
	TB

	// PB represents a petabyte (1024 terabytes).
	PB
)

// MaxHTTPBufferSize bounds the size of a single Socket.IO message (~50MB),
// large enough for media payloads while still rejecting abusive uploads.
const MaxHTTPBufferSize = 50 * MB

// StorageSize represents a data size in bytes.
type StorageSize int64

func (s StorageSize) String() string {
	switch {
	case s >= PB:
		return formatSize(s, PB, "PB")
	case s >= TB:
		return formatSize(s, TB, "TB")
	case s >= GB:
		return formatSize(s, GB, "GB")
	case s >= MB:
		return formatSize(s, MB, "MB")
	case s >= KB:
		return formatSize(s, KB, "KB")
	default:
		return formatSize(s, B, "B")
	}
}

func formatSize(size StorageSize, unit StorageSize, suffix string) string {
	return fmt.Sprintf("%.2f %s", float64(size)/float64(unit), suffix)
}

func (s StorageSize) Value() int64 {
	return int64(s)
}
