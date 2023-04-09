package pnpsqlx

type Config struct {
	DSN string `env:"DSN" envDefault:"127.0.0.1"`
}
