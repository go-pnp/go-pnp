package pnpprometheus

type Config struct {
	Addr string `env:"ADDR" envDefault:":9800"`
	Path string `env:"PATH" envDefault:"/metrics"`
}
