package pnpopentelemetry

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	fxPrivate bool

	withResourceFromSchemaURL bool
	withResourceFromDetectors bool
	resourceSchemaURL         string
	resourceAttributes        map[string]string
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithResourceFromSchemaURL(schemaURL string) optionutil.Option[options] {
	return func(o *options) {
		o.withResourceFromSchemaURL = true
		o.resourceSchemaURL = schemaURL
	}
}
