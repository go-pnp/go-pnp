package pnppromhttp

type Config struct {
	Path   string `env:"PATH" envDefault:"/metrics"`
	Method string `env:"METHOD" envDefault:"GET"`
}
