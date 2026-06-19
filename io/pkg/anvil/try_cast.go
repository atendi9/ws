package anvil

// TryCast attempts to cast an interface value to a specified type T.
// If the assertion fails, it returns the zero value of type T.
//
// Example:
//
//	val := any("hello")
//	str := anvil.TryCast[string](val) // returns "hello"
//
//	num := anvil.TryCast[int](val)    // returns 0
func TryCast[T any](val any) T {
	r, _ := val.(T)
	return r
}
