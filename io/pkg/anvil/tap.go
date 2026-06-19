package anvil

// Tap allows side-effects to be executed on a given value without modifying
// the value itself or interrupting a method chain. If the callback function
// is nil, it safely returns the original value.
//
// It returns the exact same value that was passed as the first argument.
func Tap[T any](value T, callback func(T)) T {
	if callback != nil {
		callback(value)
	}
	return value
}
