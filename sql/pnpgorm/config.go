package pnpgorm

type Config struct {
	DSN      string `env:"DSN"`
	SQLiteDB string `env:"SQLITE_PATH"`
}
