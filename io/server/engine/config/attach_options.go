package config

import (
	"time"

	"github.com/atendi9/box"
)

// AttachOptions defines the interface for configuring HTTP attachment settings.
type AttachOptions interface {
	SetPath(string)
	GetRawPath() box.Optional[string]
	Path() string

	SetDestroyUpgrade(bool)
	GetRawDestroyUpgrade() box.Optional[bool]
	DestroyUpgrade() bool

	SetDestroyUpgradeTimeout(time.Duration)
	GetRawDestroyUpgradeTimeout() box.Optional[time.Duration]
	DestroyUpgradeTimeout() time.Duration

	SetAddTrailingSlash(bool)
	GetRawAddTrailingSlash() box.Optional[bool]
	AddTrailingSlash() bool
}

// AttachOpts contains the configuration fields for the attachment options.
type AttachOpts struct {
	// path is the name of the path to capture, stored as a [box.Optional].
	path box.Optional[string]

	// destroyUpgrade indicates whether to destroy unhandled upgrade requests, stored as a [box.Optional].
	destroyUpgrade box.Optional[bool]

	// destroyUpgradeTimeout is the time in milliseconds after which unhandled requests are ended, stored as a [box.Optional].
	destroyUpgradeTimeout box.Optional[time.Duration]

	// addTrailingSlash indicates whether a trailing slash should be added to the request path, stored as a [box.Optional].
	addTrailingSlash box.Optional[bool]
}

// DefaultAttachOptions returns a new [AttachOpts] with default values.
func DefaultAttachOptions() *AttachOpts {
	return &AttachOpts{}
}

// Assign merges data from another [AttachOptions] into the current [AttachOpts].
func (a *AttachOpts) Assign(data AttachOptions) AttachOptions {
	if data == nil {
		return a
	}

	if data.GetRawPath() != nil {
		a.SetPath(data.Path())
	}

	if data.GetRawDestroyUpgradeTimeout() != nil {
		a.SetDestroyUpgradeTimeout(data.DestroyUpgradeTimeout())
	}

	if data.GetRawDestroyUpgrade() != nil {
		a.SetDestroyUpgrade(data.DestroyUpgrade())
	}

	if data.GetRawAddTrailingSlash() != nil {
		a.SetAddTrailingSlash(data.AddTrailingSlash())
	}

	return a
}

// SetPath sets the name of the path to capture.
func (a *AttachOpts) SetPath(path string) {
	a.path = box.NewSome(path)
}

// GetRawPath returns the raw name of the path to capture as a [box.Optional].
func (a *AttachOpts) GetRawPath() box.Optional[string] {
	return a.path
}

// Path returns the name of the path to capture as a [string].
func (a *AttachOpts) Path() string {
	if a.path == nil {
		return ""
	}

	return a.path.Get()
}

// SetDestroyUpgrade sets whether to destroy unhandled upgrade requests.
func (a *AttachOpts) SetDestroyUpgrade(destroyUpgrade bool) {
	a.destroyUpgrade = box.NewSome(destroyUpgrade)
}

// GetRawDestroyUpgrade returns whether to destroy unhandled upgrade requests as a [box.Optional].
func (a *AttachOpts) GetRawDestroyUpgrade() box.Optional[bool] {
	return a.destroyUpgrade
}

// DestroyUpgrade returns whether to destroy unhandled upgrade requests as a [bool].
func (a *AttachOpts) DestroyUpgrade() bool {
	if a.destroyUpgrade == nil {
		return false
	}

	return a.destroyUpgrade.Get()
}

// SetDestroyUpgradeTimeout sets the milliseconds after which unhandled requests are ended as a [time.Duration].
func (a *AttachOpts) SetDestroyUpgradeTimeout(destroyUpgradeTimeout time.Duration) {
	a.destroyUpgradeTimeout = box.NewSome(destroyUpgradeTimeout)
}

// GetRawDestroyUpgradeTimeout returns the milliseconds after which unhandled requests are ended as a [box.Optional].
func (a *AttachOpts) GetRawDestroyUpgradeTimeout() box.Optional[time.Duration] {
	return a.destroyUpgradeTimeout
}

// DestroyUpgradeTimeout returns the milliseconds after which unhandled requests are ended as a [time.Duration].
func (a *AttachOpts) DestroyUpgradeTimeout() time.Duration {
	if a.destroyUpgradeTimeout == nil {
		return 0
	}

	return a.destroyUpgradeTimeout.Get()
}

// SetAddTrailingSlash sets whether we should add a trailing slash to the request path.
func (a *AttachOpts) SetAddTrailingSlash(addTrailingSlash bool) {
	a.addTrailingSlash = box.NewSome(addTrailingSlash)
}

// GetRawAddTrailingSlash returns whether we should add a trailing slash to the request path as a [box.Optional].
func (a *AttachOpts) GetRawAddTrailingSlash() box.Optional[bool] {
	return a.addTrailingSlash
}

// AddTrailingSlash returns whether we should add a trailing slash to the request path as a [bool].
func (a *AttachOpts) AddTrailingSlash() bool {
	if a.addTrailingSlash == nil {
		return false
	}

	return a.addTrailingSlash.Get()
}
