package pnpprometheus

type Config struct {
	Port int    `env:"PORT,default=9800"`
	Path string `env:"PATH,default=/metrics"`
}
