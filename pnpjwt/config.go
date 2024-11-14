package pnpjwt

type Config struct {
	SigningMethod string `env:"SIGNING_METHOD,notEmpty"`
	SigningKey    string `env:"SIGNING_KEY,notEmpty"`
}
