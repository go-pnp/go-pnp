module github.com/go-pnp/go-pnp/http/pnphttpserverrecovery

go 1.22.0

require (
	github.com/go-pnp/go-pnp v1.1.3
	github.com/go-pnp/go-pnp/http/pnphttpserver v0.0.13
	github.com/gorilla/mux v1.8.1
	github.com/stretchr/testify v1.11.1
	go.uber.org/fx v1.24.0
)

require (
	github.com/caarlos0/env/v10 v10.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-pnp/go-pnp => ../..
