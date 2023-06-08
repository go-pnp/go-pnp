package pnphttphealthcheck

type Config struct {
	Path   string `env:"PATH" envDefault:"/health"`
	Method string `env:"METHOD" envDefault:"GET"`
}
