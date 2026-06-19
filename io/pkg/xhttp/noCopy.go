package xhttp

// noCopy may be embedded into structs which must not be copied
// after the first use.
type noCopy struct{}

// Lock is a no-op used by go vet to check for copying.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
