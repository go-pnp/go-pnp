package pnpgorm

type Config struct {
	DSN string `env:"DSN,notEmpty"`
}

type SQLiteConfig struct {
	Path string `env:"SQLITE_PATH,notEmpty"`
}
