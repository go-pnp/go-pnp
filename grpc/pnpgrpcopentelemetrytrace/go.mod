module github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetrytrace

go 1.22

require (
	github.com/go-pnp/go-pnp v1.0.0
	github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetry v0.0.13
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.55.0
	go.opentelemetry.io/otel/trace v1.30.0
	go.uber.org/fx v1.22.2
	google.golang.org/grpc v1.66.2
)

require (
	github.com/caarlos0/env/v10 v10.0.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-pnp/go-pnp/grpc/pnpgrpcclient v0.1.2 // indirect
	github.com/go-pnp/go-pnp/grpc/pnpgrpcserver v0.0.9 // indirect
	github.com/go-pnp/jobber v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	go.opentelemetry.io/otel v1.30.0 // indirect
	go.opentelemetry.io/otel/metric v1.30.0 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace (
	github.com/go-pnp/go-pnp/grpc/pnpgrpcclient => ../pnpgrpcclient
	github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetry => ../pnpgrpcopentelemetry
	github.com/go-pnp/go-pnp/grpc/pnpgrpcserver => ../pnpgrpcserver
)
