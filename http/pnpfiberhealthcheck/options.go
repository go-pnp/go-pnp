package pnpfiberhealthcheck

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-pnp/go-pnp/pkg/optionutil"
)

type options struct {
	path           string
	method         string
	fxPrivate      bool
	responseWriter func(alive bool, checkResults map[string]error, ctx *fiber.Ctx)
}

func newOptions(opts []optionutil.Option[options]) *options {
	return optionutil.ApplyOptions(&options{
		path:           "/health",
		method:         "GET",
		responseWriter: WriteResponse,
	}, opts...)
}

// WithEndpoint sets the path and method for the healthcheck endpoint.
func WithEndpoint(method, path string) optionutil.Option[options] {
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

func WithResponseWriter(writer func(alive bool, checkResults map[string]error, ctx *fiber.Ctx)) optionutil.Option[options] {
	return func(o *options) {
		o.responseWriter = writer
	}
}
