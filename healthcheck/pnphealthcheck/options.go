package pnphealthcheck

import (
	"time"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	// all provides are private
	fxPrivate bool

	// default checks timeout
	defaultTimeout time.Duration
}

func newOptions(opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		fxPrivate:      false,
		defaultTimeout: time.Second * 2,
	}, opts...)
}

func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithChecksTimeout(timeout time.Duration) optionutil.Option[options] {
	return func(o *options) {
		o.defaultTimeout = timeout
	}
}
