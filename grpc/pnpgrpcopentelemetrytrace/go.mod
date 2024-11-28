module github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetrytrace

go 1.22.7

toolchain go1.23.1

require (
	github.com/go-pnp/go-pnp v1.1.3
	github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetry v0.0.15
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.57.0
	go.opentelemetry.io/otel/trace v1.32.0
	go.uber.org/fx v1.23.0
	google.golang.org/grpc v1.68.0
)

require (
	github.com/caarlos0/env/v10 v10.0.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-pnp/go-pnp/grpc/pnpgrpcclient v0.1.4 // indirect
	github.com/go-pnp/go-pnp/grpc/pnpgrpcserver v0.0.11 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.uber.org/dig v1.18.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241118233622-e639e219e697 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

replace (
	github.com/go-pnp/go-pnp/grpc/pnpgrpcclient => ../pnpgrpcclient
	github.com/go-pnp/go-pnp/grpc/pnpgrpcopentelemetry => ../pnpgrpcopentelemetry
	github.com/go-pnp/go-pnp/grpc/pnpgrpcserver => ../pnpgrpcserver
)
