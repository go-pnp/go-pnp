package pnpsentry

import "github.com/getsentry/sentry-go"

type Config struct {
	// The DSN to use. If the DSN is not set, the client is effectively
	// disabled.
	DSN string `env:"DSN"`
	// In debug mode, the debug information is printed to stdout to help you
	// understand what sentry is doing.
	Debug bool `env:"DEBUG"`
	// Configures whether SDK should generate and attach stacktraces to pure
	// capture message calls.
	AttachStacktrace bool `env:"ATTACH_STACKTRACE"`
	// The sample rate for event submission in the range [0.0, 1.0]. By default,
	// all events are sent. Thus, as a historical special case, the sample rate
	// 0.0 is treated as if it was 1.0. To drop all events, set the DSN to the
	// empty string.
	SampleRate float64 `env:"SAMPLE_RATE" envDefault:"1"`
	// Enable performance tracing.
	EnableTracing bool `env:"ENABLE_TRACING"`
	// The sample rate for sampling traces in the range [0.0, 1.0].
	TracesSampleRate float64 `env:"TRACES_SAMPLE_RATE" envDefault:"1"`
	// Used to customize the sampling of traces, overrides TracesSampleRate.
	TracesSampler sentry.TracesSampler
	// The server name to be reported.
	ServerName string `env:""`
	// The release to be sent with events.
	//
	// Some Sentry features are built around releases, and, thus, reporting
	// events with a non-empty release improves the product experience. See
	// https://docs.sentry.io/product/releases/.
	//
	// If Release is not set, the SDK will try to derive a default value
	// from environment variables or the Git repository in the working
	// directory.
	//
	// If you distribute a compiled binary, it is recommended to set the
	// Release value explicitly at build time. As an example, you can use:
	//
	// 	go build -ldflags='-X main.release=VALUE'
	//
	// That will set the value of a predeclared variable 'release' in the
	// 'main' package to 'VALUE'. Then, use that variable when initializing
	// the SDK:
	//
	// 	sentry.Init(ClientOptions{Release: release})
	//
	// See https://golang.org/cmd/go/ and https://golang.org/cmd/link/ for
	// the official documentation of -ldflags and -X, respectively.
	Release string `env:"RELEASE"`
	// The dist to be sent with events.
	Dist string `env:"DIST"`
	// The environment to be sent with events.
	Environment string `env:"ENVIRONMENT"`
	// Maximum number of breadcrumbs
	// when MaxBreadcrumbs is negative then ignore breadcrumbs.
	MaxBreadcrumbs int `env:"MAX_BREADCRUMBS"`
	// Maximum number of spans.
	//
	// See https://develop.sentry.dev/sdk/envelopes/#size-limits for size limits
	// applied during event ingestion. Events that exceed these limits might get dropped.
	MaxSpans int `env:"MAX_SPANS"`
}
