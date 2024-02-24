package pnpgorm

type Config struct {
	DSN      string `env:"DSN,notEmpty"`
	SQLiteDB string `env:"SQLITE_PATH"`
}
