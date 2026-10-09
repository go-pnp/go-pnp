package panicutil

import "github.com/pkg/errors"

// NewError keeps the panic-site stack but hides an error panic value from errors.Is and errors.As,
// so a panic is never mistaken for the handled error it carries.
func NewError(panicValue any) error {
	return errors.Errorf("panic: %v", panicValue)
}

// Recover must be deferred directly (defer panicutil.Recover(...)): recover only stops a panic
// when it is called by the deferred function itself.
func Recover(handleError func(err error)) {
	if panicValue := recover(); panicValue != nil {
		handleError(NewError(panicValue))
	}
}

func Call(fn func() error) (err error) {
	defer Recover(func(panicError error) {
		err = panicError
	})

	return fn()
}
