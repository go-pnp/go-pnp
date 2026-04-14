module github.com/go-pnp/go-pnp/connectrpc/pnpsentryconnectrpchandling/example

go 1.25.1

replace (
	github.com/go-pnp/go-pnp/connectrpc/pnpsentryconnectrpchandling => ../
	github.com/go-pnp/go-pnp/http/pnphttpserversentry => ../../../http/pnphttpserversentry
)

require (
	connectrpc.com/connect v1.19.1
	github.com/go-pnp/go-pnp v1.1.4
	github.com/go-pnp/go-pnp/connectrpc/pnpconnectrpchandling v0.0.6
	github.com/go-pnp/go-pnp/connectrpc/pnpsentryconnectrpchandling v0.0.0-00010101000000-000000000000
	github.com/go-pnp/go-pnp/http/pnphttpserver v0.0.14
	github.com/go-pnp/go-pnp/http/pnphttpserversentry v0.0.4
	github.com/go-pnp/go-pnp/logging/pnpzap v0.0.16
	github.com/go-pnp/go-pnp/logging/pnpzapsentry v0.0.5
	github.com/go-pnp/go-pnp/pnpenv v1.0.4
	github.com/go-pnp/go-pnp/pnpsentry v0.0.4
	go.uber.org/fx v1.24.0
)

require (
	github.com/caarlos0/env/v10 v10.0.0 // indirect
	github.com/getsentry/sentry-go v0.31.1 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)
