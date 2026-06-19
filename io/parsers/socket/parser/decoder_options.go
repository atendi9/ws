package parser

import (
	"github.com/atendi9/box"
)

// DecoderOptions holds configuration options for the Decoder.
// It provides methods to configure values using [box.Optional].
type DecoderOptions interface {
	SetMaxAttachments(uint64)
	GetRawMaxAttachments() box.Optional[uint64]
	MaxAttachments() uint64

	SetMaxNamespaceLength(int)
	GetRawMaxNamespaceLength() box.Optional[int]
	MaxNamespaceLength() int

	SetMaxPacketIDLength(int)
	GetRawMaxPacketIDLength() box.Optional[int]
	MaxPacketIDLength() int
}

// DecoderOpts implements [DecoderOptions] to manage configuration options.
type DecoderOpts struct {
	// maxAttachments is the maximum number of binary attachments allowed per packet.
	// Defaults to DefaultMaxAttachments (10) if not set or set to 0.
	maxAttachments box.Optional[uint64]

	// maxNamespaceLength is the maximum allowed length of a namespace name.
	// Defaults to DefaultMaxNamespaceLength (512) if not set or set to 0.
	maxNamespaceLength box.Optional[int]

	// maxPacketIDLength is the maximum allowed length of a packet ID string.
	// Defaults to DefaultMaxPacketIDLength (20) if not set or set to 0.
	maxPacketIDLength box.Optional[int]
}

// DefaultDecoderOptions returns a new instance of [DecoderOpts] with no options set.
func DefaultDecoderOptions() *DecoderOpts {
	return &DecoderOpts{}
}

// Assign merges the provided [DecoderOptions] into the current [DecoderOpts].
// If the provided data is nil, it returns the original [DecoderOpts].
func (d *DecoderOpts) Assign(data DecoderOptions) *DecoderOpts {
	if data == nil {
		return d
	}

	if data.GetRawMaxAttachments() != nil {
		d.SetMaxAttachments(data.MaxAttachments())
	}

	if data.GetRawMaxNamespaceLength() != nil {
		d.SetMaxNamespaceLength(data.MaxNamespaceLength())
	}

	if data.GetRawMaxPacketIDLength() != nil {
		d.SetMaxPacketIDLength(data.MaxPacketIDLength())
	}

	return d
}

// SetMaxAttachments sets the maximum number of binary attachments allowed per packet.
func (d *DecoderOpts) SetMaxAttachments(maxAttachments uint64) {
	d.maxAttachments = box.NewSome(maxAttachments)
}

// GetRawMaxAttachments returns the raw [box.Optional] wrapper for the max attachments value.
func (d *DecoderOpts) GetRawMaxAttachments() box.Optional[uint64] {
	return d.maxAttachments
}

// MaxAttachments returns the maximum number of attachments, or 0 if not set.
func (d *DecoderOpts) MaxAttachments() uint64 {
	if d.maxAttachments == nil {
		return 0
	}
	return d.maxAttachments.Get()
}

// SetMaxNamespaceLength sets the maximum allowed length of a namespace name.
func (d *DecoderOpts) SetMaxNamespaceLength(maxNamespaceLength int) {
	d.maxNamespaceLength = box.NewSome(maxNamespaceLength)
}

// GetRawMaxNamespaceLength returns the raw [box.Optional] wrapper for the max namespace length value.
func (d *DecoderOpts) GetRawMaxNamespaceLength() box.Optional[int] {
	return d.maxNamespaceLength
}

// MaxNamespaceLength returns the maximum allowed length of a namespace, or 0 if not set.
func (d *DecoderOpts) MaxNamespaceLength() int {
	if d.maxNamespaceLength == nil {
		return 0
	}
	return d.maxNamespaceLength.Get()
}

// SetMaxPacketIDLength sets the maximum allowed length of a packet ID string.
func (d *DecoderOpts) SetMaxPacketIDLength(maxPacketIDLength int) {
	d.maxPacketIDLength = box.NewSome(maxPacketIDLength)
}

// GetRawMaxPacketIDLength returns the raw [box.Optional] wrapper for the max packet ID length value.
func (d *DecoderOpts) GetRawMaxPacketIDLength() box.Optional[int] {
	return d.maxPacketIDLength
}

// MaxPacketIDLength returns the maximum allowed length of a packet ID, or 0 if not set.
func (d *DecoderOpts) MaxPacketIDLength() int {
	if d.maxPacketIDLength == nil {
		return 0
	}
	return d.maxPacketIDLength.Get()
}
