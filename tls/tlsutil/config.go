package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type ClientTLSConfig struct {
	Enabled               bool   `env:"ENABLED" envDefault:"false"`
	CertPath              string `env:"CERT_PATH"`
	KeyPath               string `env:"KEY_PATH"`
	InsecureSkipVerify    bool   `env:"INSECURE_SKIP_VERIFY" envDefault:"false"`
	RootCAs               string `env:"ROOT_CA_PATH"`
	AppendSystemCAsToRoot bool   `env:"APPEND_SYSTEM_CAS_TO_ROOT" envDefault:"false"`
}

func (g *ClientTLSConfig) TLSConfig() (*tls.Config, error) {
	if g == nil || g.Enabled == false {
		return nil, nil
	}

	result := &tls.Config{
		InsecureSkipVerify: g.InsecureSkipVerify,
	}

	clientCert, err := tls.LoadX509KeyPair(g.CertPath, g.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load tls key pair: %w", err)
	}

	result.Certificates = []tls.Certificate{clientCert}

	if g.RootCAs != "" {
		pool, err := loadX509CertPool(g.RootCAs, g.AppendSystemCAsToRoot)
		if err != nil {
			return nil, fmt.Errorf("load root CAs: %w", err)
		}
		result.RootCAs = pool
	}

	return result, nil
}

type ServerTLSConfig struct {
	Enabled                 bool   `env:"ENABLED" envDefault:"false"`
	CertPath                string `env:"CERT_PATH"`
	KeyPath                 string `env:"KEY_PATH"`
	ClientAuth              string `env:"CLIENT_AUTH"`
	ClientCAs               string `env:"CLIENT_CA_PATH"`
	AppendSystemCAsToClient bool   `env:"APPEND_SYSTEM_CAS_TO_CLIENT" envDefault:"false"`
}

func (g *ServerTLSConfig) TLSConfig() (*tls.Config, error) {
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
			return nil, fmt.Errorf("unknown client auth strategy: %s. valid values are: [no_client_cert, request_client_cert, require_any_client_cert, verify_client_cert_if_given, require_and_verify_client_cert]", g.ClientAuth)
		}
	}

	serverCert, err := tls.LoadX509KeyPair(g.CertPath, g.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load tls key pair: %w", err)
	}

	result.Certificates = []tls.Certificate{serverCert}

	if g.ClientCAs != "" {
		pool, err := loadX509CertPool(g.ClientCAs, g.AppendSystemCAsToClient)
		if err != nil {
			return nil, fmt.Errorf(" oad root CAs: %w", err)
		}
		result.ClientCAs = pool
	}

	return result, nil
}

func loadX509CertPool(paths string, appendSystem bool) (*x509.CertPool, error) {
	var pool *x509.CertPool
	if appendSystem {
		var err error
		pool, err = x509.SystemCertPool()
		if err != nil {
			return nil, errors.Wrap(err, "can't load system x509 cert pool")
		}
	} else {
		pool = x509.NewCertPool()
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
