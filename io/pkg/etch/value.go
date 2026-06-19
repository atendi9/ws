package etch

func Value(
	value string,
	_default string,
) string {
	if len(value) == 0 {
		return _default
	}
	return value
}
