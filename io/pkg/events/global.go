package events

// defaultEmitter acts as the package-level singleton broker instance.
var defaultEmitter = NewEmitter()

// AddListener registers multiple [Listener] callbacks to the package-level default emitter for a given [Name].
func AddListener(evt Name, listeners ...Listener) error {
	return defaultEmitter.AddListener(evt, listeners...)
}

// On acts as an alias shortcut for [AddListener], attaching callbacks to a package-level [Name].
func On(evt Name, listeners ...Listener) error {
	return AddListener(evt, listeners...)
}

// Emit broadcasts the specified [Name] with arbitrary payloads to all handlers tied to the global singleton.
func Emit(evt Name, data ...any) {
	defaultEmitter.Emit(evt, data...)
}

// Names lists every unique [Name] category presently registered within the global instance.
func Names() []Name {
	return defaultEmitter.Names()
}

// ListenerCount calculates the active total number of handlers appended to the chosen global [Name].
func ListenerCount(evt Name) int {
	return defaultEmitter.ListenerCount(evt)
}

// Listeners reads and lists all active internal [Listener] functions assigned to the given [Name].
func Listeners(evt Name) []Listener {
	return defaultEmitter.Listeners(evt)
}

// Once hooks an ephemeral [Listener] callback that executes exactly once and unregisters from the default system.
func Once(evt Name, listeners ...Listener) error {
	return defaultEmitter.Once(evt, listeners...)
}

// RemoveListener performs pointer comparisons to discard a specific [Listener] hook from a global [Name].
func RemoveListener(evt Name, listener Listener) bool {
	return defaultEmitter.RemoveListener(evt, listener)
}

// RemoveAllListeners completely flushes out all subscription instances under a targeted [Name].
func RemoveAllListeners(evt Name) bool {
	return defaultEmitter.RemoveAllListeners(evt)
}

// Clear wipes all categories and data inside the internal package-level event registration mapping.
func Clear() {
	defaultEmitter.Clear()
}

// Len gives out the unique amount of active structural event name categories inside the package singleton.
func Len() int {
	return defaultEmitter.Len()
}
