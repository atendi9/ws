package io

// ParamsDecoder converts a map of parameters into a specified generic type.
// It serializes the provided map into JSON data and then deserializes it into the target structure.
// In case of a type mismatch during the decoding process, it may return an [json.InvalidUnmarshalError]
// or an [json.UnmarshalTypeError].
func ParamsDecoder[Params any](
	params []byte,
	decoder func([]byte) (Params, error),
) (output Params, err error) {
	output, err = decoder(params)
	return
}
