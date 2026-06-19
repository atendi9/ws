package anvil

func New[T any](v T) *T {
	return new(v)
}
