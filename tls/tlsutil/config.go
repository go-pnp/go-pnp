package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type TLSConfig struct {
	Enabled     bool   `env:"ENABLED" envDefault:"false"`
	TLSCertPath string `env:"CERT_PATH"`
	TLSKeyPath  string `env:"KEY_PATH"`
	ClientAuth  string `env:"CLIENT_AUTH"`
	ClientCAs   string `env:"CLIENT_CA_PATHS"`
}

func (g *TLSConfig) TLSConfig() (*tls.Config, error) {
	if g == nil || g.Enabled == false {
		return nil, nil
	}

	result := &tls.Config{
		ClientAuth: tls.NoClientCert,
	}

	if g.ClientAuth != "" {
		switch g.ClientAuth {
		case "request_client_cert":
			result.ClientAuth = tls.RequestClientCert
		case "require_any_client_cert":
			result.ClientAuth = tls.RequireAnyClientCert
		case "verify_client_cert_if_given":
			result.ClientAuth = tls.VerifyClientCertIfGiven
		case "require_and_verify_client_cert":
			result.ClientAuth = tls.RequireAndVerifyClientCert
		default:
			return nil, errors.Errorf("unknown client auth strategy: %s. valid values are: [no_client_cert, request_client_cert, require_any_client_cert, verify_client_cert_if_given, require_and_verify_client_cert]", g.ClientAuth)
		}
	}

	serverCert, err := tls.LoadX509KeyPair(g.TLSCertPath, g.TLSKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "can't load tls key pair")
	}

	result.Certificates = []tls.Certificate{serverCert}

	if g.ClientCAs != "" {
		pool, err := loadX509CertPool(g.ClientCAs)
		if err != nil {
			return nil, errors.Wrap(err, "can't load root CAs")
		}
		result.ClientCAs = pool
	}

	return result, nil
}

func loadX509CertPool(paths string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, errors.Wrap(err, "can't load system x509 cert pool")
	}
	for _, path := range strings.Split(paths, ";") {
		rootPEM, err := os.ReadFile(path)
		if err != nil || rootPEM == nil {
			return nil, fmt.Errorf("nats: error loading or parsing rootCA file %v: %v", paths, err)
		}

		ok := pool.AppendCertsFromPEM(rootPEM)
		if !ok {
			return nil, fmt.Errorf("nats: failed to parse root certificate from %q", path)
		}
	}

	return pool, nil
}
