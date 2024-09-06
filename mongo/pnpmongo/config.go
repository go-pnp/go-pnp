package pnpmongo

type Config struct {
	DSN string `env:"DSN,notEmpty"`
}
