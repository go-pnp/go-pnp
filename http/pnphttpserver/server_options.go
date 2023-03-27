package pnphttpserver

type serverOptions struct {
	start      bool
	provideMux bool
	config     *Config
}

type ServerOption func(*serverOptions)

func WithConfig(config *Config) ServerOption {
	return func(o *serverOptions) {
		o.config = config
	}
}

func DoNotProvideMux() ServerOption {
	return func(o *serverOptions) {
		o.provideMux = false
	}
}

func DoNotStart() ServerOption {
	return func(o *serverOptions) {
		o.start = false
	}
}
