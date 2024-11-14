package pnphttpservercors

type Config struct {
	AllowAll           bool     `env:"ALLOW_ALL_ORIGINS" envDefault:"false"`
	AllowedHeaders     []string `env:"ALLOWED_HEADERS" envDefault:"*"`
	AllowedOriginGlobs []string `env:"ALLOWED_ORIGIN_GLOBS" envDefault:""`
	AllowedOrigins     []string `env:"ALLOWED_ORIGINS" envDefault:""`
}
