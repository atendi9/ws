package config

// Options defines the combined interface for both [AttachOptions] and [ServerOptions].
type Options interface {
	AttachOptions
	ServerOptions
}

// Opts combines [AttachOpts] and [ServerOpts] into a single configuration structure.
type Opts struct {
	*AttachOpts
	*ServerOpts
}

// DefaultOptions returns a new [Opts] with default values for both attach and server options.
func DefaultOptions() *Opts {
	return &Opts{
		AttachOpts: DefaultAttachOptions(),
		ServerOpts: DefaultServerOptions(),
	}
}

// Assign merges data from another [Options] into the current [Opts].
func (o *Opts) Assign(data Options) Options {
	if data == nil {
		return o
	}

	o.AttachOpts.Assign(data)
	o.ServerOpts.Assign(data)

	return o
}
