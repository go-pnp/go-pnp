module github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetrytrace

go 1.22

require (
	github.com/go-pnp/go-pnp v1.0.0
	github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetry v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.48.0
	go.opentelemetry.io/otel/trace v1.23.1
	go.uber.org/fx v1.22.2
	google.golang.org/grpc v1.62.0
)

require (
	github.com/caarlos0/env/v10 v10.0.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-pnp/go-pnp/grpc/pnpgrpcclient v0.1.1 // indirect
	github.com/go-pnp/go-pnp/grpc/pnpgrpcserver v0.0.7 // indirect
	github.com/go-pnp/jobber v1.0.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	go.opentelemetry.io/otel v1.23.1 // indirect
	go.opentelemetry.io/otel/metric v1.23.1 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240221002015-b0ce06bbee7c // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace (
	github.com/go-pnp/go-pnp/grpc/pnpgrpcclient => ../pnpgrpcclient
	github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetry => ../pnpgrpcopentelemetry
	github.com/go-pnp/go-pnp/grpc/pnpgrpcserver => ../pnpgrpcserver
)
