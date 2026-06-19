package anvil

// Key represents a comparable type that can be used as a key in a map.
type Key = comparable

// Input represents any type that is used as an input value in map transformations.
type Input = any

// NewInput creates a new Input from the provided value.
func NewInput(v any) Input {
	return Input(v)
}

// Output represents any type that is used as an output value in map transformations.
type Output = any

// NewOutput creates a new Output from the provided value.
func NewOutput(v any) Output {
	return Output(v)
}

// Values transforms the values of a map of type map[K]I into a new
// map of type map[K]O by applying the provided transformation
// function to each value. If the input map is nil, it returns nil.
func Values[K Key, I Input, O Output](
	input map[K]I,
	transform func(I) O,
) map[K]O {
	if input == nil {
		return nil
	}

	result := make(map[K]O, len(input))
	for key, value := range input {
		result[key] = transform(value)
	}
	return result
}
