module github.com/go-pnp/go-pnp/opentelemetry/pnpopentelemetry

go 1.22

require (
	github.com/go-pnp/go-pnp v1.1.2
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/sdk v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
	go.uber.org/fx v1.23.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
)

replace cloud.google.com/go/compute/metadata => cloud.google.com/go v0.65.0
