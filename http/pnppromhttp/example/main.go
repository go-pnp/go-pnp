package main

import (
	"github.com/go-pnp/go-pnp/http/pnphttpserver"
	"github.com/go-pnp/go-pnp/prometheus/pnpprometheus"
	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/http/pnppromhttp"
)

func main() {
	fx.New(
		pnpprometheus.Module(),
		pnphttpserver.Module(),
		pnppromhttp.Module(
			pnppromhttp.WithEndpoint("GET", "/metrics"), // not required
			pnppromhttp.RegisterInMux(true),             // not required
		),
	).Run()
}
