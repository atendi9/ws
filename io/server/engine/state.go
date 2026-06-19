package engine

import "github.com/atendi9/ws/io/pkg/state"

var (
	// StateOpening represents the server engine when it is in the process of opening.
	StateOpening = NewServerState("opening")

	// StateOpen represents the server engine when it is fully open and ready to accept connections.
	StateOpen = NewServerState("open")

	// StateClosed represents the server engine when it is fully closed and inactive.
	StateClosed = NewServerState("closed")

	// StateClosing represents the server engine when it is in the process of terminating its operations.
	StateClosing = NewServerState("closing")
)

// NewServerState creates and returns a new [state.State] specific to the server engine
// using the provided string identifier.
func NewServerState(s string) state.State {
	return state.NewState("server-engine", s)
}
