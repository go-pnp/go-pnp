package optionutil

// Option is a function optional parameter.
type Option[T any] func(*T)

func ApplyOptions[T any](val *T, opts ...Option[T]) *T {
	for _, opt := range opts {
		opt(val)
	}

	return val
}
