package pnpgorm

type Config struct {
	DSN string `env:"DSN,notEmpty"`
}
