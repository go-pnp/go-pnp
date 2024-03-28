package main

import (
	"fmt"
	"os"

	"go.uber.org/fx"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
)

type Config struct {
	Foo string `env:"FOO,notEmpty"`
	Bar struct {
		Baz string `env:"BAZ,notEmpty" envDefault:"baz"`
	} `envPrefix:"BAR_"`
}

func main() {
	os.Setenv("TST_FOO", "foo")
	os.Setenv("TST_BAR_BAZ", "baz")
	fx.New(
		fx.Provide(
			configutil.NewPrefixedConfigProvider[Config]("TST_"),
			configutil.NewPrefixedConfigInfoProvider[Config]("TST_"),
			configutil.NewPrefixedConfigProvider[fxutil.Config]("TST1_"),
			configutil.NewPrefixedConfigInfoProvider[fxutil.Config]("TST1_"),
		),
		// Print info about all provided configs
		fx.Invoke(func(params configutil.ConfigsInfoIn) {
			for _, configInfo := range params.ConfigsInfo {
				fmt.Printf("# %s\n", configInfo.ConfigType.String())
				for _, field := range configInfo.Fields {
					fmt.Println(field.Key, field.DefaultValue, field.Required, field.NotEmpty)
				}
			}
		}),
		// Get config
		fx.Invoke(func(config *Config) {
			fmt.Println(config.Foo)
			fmt.Println(config.Bar.Baz)
		}),
		// Dump config fields in .env format
		fx.Invoke(func(params configutil.ConfigsInfoIn) {
			configutil.DumpConfigsInDotEnvFormat(params)
		}),
	)
}
