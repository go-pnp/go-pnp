package pnphttpservercors

type Config struct {
	AllowAll            bool     `env:"ALLOW_ALL_ORIGINS" envDefault:"false"`
	AllowedHeaders      []string `env:"ALLOWED_HEADERS" envDefault:"*"`
	AllowedOriginGlobs  []string `env:"ALLOWED_ORIGIN_GLOBS" envDefault:""`
	AllowedOrigins      []string `env:"ALLOWED_ORIGINS" envDefault:""`
	AllowedMethods      []string `env:"ALLOWED_METHODS" envDefault:""`
	ExposedHeaders      []string `env:"EXPOSED_HEADERS" envDefault:""`
	AllowCredentials    bool     `env:"ALLOW_CREDENTIALS" envDefault:"false"`
	AllowPrivateNetwork bool     `env:"ALLOW_PRIVATE_NETWORK" envDefault:"false"`
	MaxAge              int      `env:"MAX_AGE" envDefault:"0"`
}
