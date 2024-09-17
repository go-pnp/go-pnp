package pnphealthcheckgorm

import (
	"time"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	// all provides to fx container are private
	fxPrivate bool

	// checks timeout
	timeout time.Duration

	// health checker name
	name string
}

func newOptions(opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		name:      "gorm",
		fxPrivate: false,
		timeout:   time.Second * 2,
	}, opts...)
}

func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithTimeout(timeout time.Duration) optionutil.Option[options] {
	return func(o *options) {
		o.timeout = timeout
	}
}

func WithName(name string) optionutil.Option[options] {
	return func(o *options) {
		o.name = name
	}
}
