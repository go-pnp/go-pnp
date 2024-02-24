package pnphttphealthcheck

import (
	"net/http"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	path                string
	method              string
	fxPrivate           bool
	configFromContainer bool
	responseWriter      func(alive bool, checkResults map[string]error, w http.ResponseWriter)
	envConfigPrefix     string
}

func newOptions(opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		path:            "/health",
		method:          "GET",
		responseWriter:  WriteResponse,
		envConfigPrefix: "HTTP_HEALTHCHECK_",
	}, opts...)
}

// WithEndpoint sets the path and method for the healthcheck endpoint.
func WithEndpoint(path, method string) optionutil.Option[options] {
	return func(o *options) {
		o.path = path
		o.method = method
	}
}

// WithFxPrivate is an option to add fx.Private to all module provides.
func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithConfigFromContainer() optionutil.Option[options] {
	return func(o *options) {
		o.configFromContainer = true
	}
}

func WithEnvConfigPrefix(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.envConfigPrefix = prefix
	}
}

func WithResponseWriter(writer func(alive bool, checkResults map[string]error, w http.ResponseWriter)) optionutil.Option[options] {
	return func(o *options) {
		o.responseWriter = writer
	}
}
