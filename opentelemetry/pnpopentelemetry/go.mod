module github.com/go-pnp/go-pnp/opentelemetry/pnpopentelemetry

go 1.22

require (
	github.com/go-pnp/go-pnp v0.0.12
	go.opentelemetry.io/otel v1.23.1
	go.opentelemetry.io/otel/sdk v1.23.1
	go.opentelemetry.io/otel/trace v1.23.1
	go.uber.org/fx v1.20.1
)

require (
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-pnp/jobber v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	go.opentelemetry.io/otel/metric v1.23.1 // indirect
	go.uber.org/dig v1.17.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

replace cloud.google.com/go/compute/metadata => cloud.google.com/go v0.65.0
