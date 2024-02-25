package pnpfiber

import (
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/gofiber/fiber/v2"
)

type options struct {
	// all provides are private
	fxPrivate bool

	// if true, the config will be consumed from outside
	configFromContainer bool

	// if provided, module will load config from env variables with this prefix
	// default is HTTP_SERVER_
	configPrefix string

	// if true, the fiber config will be consumed from outside
	fiberConfigFromContainer bool

	// fiber config to be used
	fiberConfig *fiber.Config

	// if true, the server will be started on app start
	startServer bool
}

func newOptions() *options {
	return &options{
		fxPrivate:           false,
		configFromContainer: false,
		configPrefix:        "HTTP_SERVER_",
		fiberConfig:         nil,
		startServer:         true,
	}
}

func WithFxPrivate() optionutil.Option[options] {
	return func(o *options) {
		o.fxPrivate = true
	}
}

func WithConfigPrefix(prefix string) optionutil.Option[options] {
	return func(o *options) {
		o.configPrefix = prefix
	}
}

func WithFiberConfig(config fiber.Config) optionutil.Option[options] {
	return func(o *options) {
		o.fiberConfig = &config
	}
}

func WithStartServer(startServer bool) optionutil.Option[options] {
	return func(o *options) {
		o.startServer = startServer
	}
}
