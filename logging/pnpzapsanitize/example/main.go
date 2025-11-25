package main

import (
	"context"
	"os"

	"github.com/go-pnp/go-pnp/logging"
	"github.com/go-pnp/go-pnp/logging/pnpzap"
	"github.com/go-pnp/go-pnp/logging/pnpzapsanitize"
	"go.uber.org/fx"
)

type StructExample struct {
	Token  string
	Normal string
}

func main() {
	os.Setenv("ENVIRONMENT", "dev")
	fx.New(
		pnpzap.Module(),
		pnpzapsanitize.Module(),
		fx.Invoke(func(logger *logging.Logger) {
			logger.WithFields(map[string]any{
				"password": "mysecretpassword",
				"token":    "mysecrettoken",
				"normal":   "normalvalue",
				"struct":   StructExample{Token: "structsecrettoken", Normal: "normalvalue"},
				"mapexample": map[string]string{
					"api_key": "mysecretapikey",
					"other":   "othervalue",
				},
			}).Info(context.Background(), "Hello from test logger")
		}),
	).Run()
}
