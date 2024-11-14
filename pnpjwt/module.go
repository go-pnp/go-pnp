package pnpjwt

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/go-pnp/go-pnp/config/configutil"
	"github.com/go-pnp/go-pnp/fxutil"
	"github.com/go-pnp/go-pnp/pkg/optionutil"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/fx"
)

type SignParams struct {
	Method     jwt.SigningMethod
	SigningKey any
}

// Module is an fx module that provides SignParams to fx DI container.
func Module(opts ...optionutil.Option[options]) fx.Option {
	options := newOptions(opts...)

	moduleBuilder := &fxutil.OptionsBuilder{
		PrivateProvides: options.fxPrivate,
	}
	moduleBuilder.ProvideIf(!options.configFromContainer,
		configutil.NewPrefixedConfigProvider[Config](options.configPrefix),
		configutil.NewPrefixedConfigInfoProvider[Config](options.configPrefix),
	)
	moduleBuilder.Provide(newJWTTokensManager)

	return moduleBuilder.Build()

}

func newJWTTokensManager(config *Config) (*SignParams, error) {
	signingMethod := jwt.GetSigningMethod(config.SigningMethod)
	if signingMethod == nil {
		return nil, fmt.Errorf("unsupported signing method: %s", config.SigningMethod)
	}

	var key any
	switch config.SigningMethod {
	case "HS256", "HS384", "HS512":
		key = []byte(config.SigningKey)
	case "RS256", "RS384", "RS512":
		privateKey, err := loadRSAPrivateKey(config.SigningKey)
		if err != nil {
			return nil, fmt.Errorf("error loading RSA private key: %w", err)
		}
		key = privateKey
	case "ES256", "ES384", "ES512":
		privateKey, err := loadECDSAPrivateKey(config.SigningKey)
		if err != nil {
			return nil, fmt.Errorf("error loading ECDSA private key: %w", err)
		}
		key = privateKey
	case "PS256", "PS384", "PS512":
		privateKey, err := loadRSAPrivateKey(config.SigningKey)
		if err != nil {
			return nil, fmt.Errorf("error loading RSA private key: %w", err)
		}
		key = privateKey
	case "EdDSA":
		privateKey, err := loadEdDSAPrivateKey(config.SigningKey)
		if err != nil {
			return nil, fmt.Errorf("error loading EdDSA private key: %w", err)
		}
		key = privateKey
	default:
		return nil, fmt.Errorf("unsupported signing method: %s", config.SigningMethod)
	}

	return &SignParams{Method: signingMethod, SigningKey: key}, nil
}

func loadRSAPrivateKey(secret string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(secret))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func loadECDSAPrivateKey(secret string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(secret))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

func loadEdDSAPrivateKey(secret string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(secret))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}
	return block.Bytes, nil
}
