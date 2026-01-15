package control

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

func LoadClientCredentials(certFile, keyFile, caFile string, skipVerify bool) (credentials.TransportCredentials, error) {
	if certFile == "" || keyFile == "" {
		if skipVerify {
			return credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}), nil
		}
		return nil, fmt.Errorf("certificate and key files are required")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("loading key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: skipVerify,
	}

	if caFile != "" && !skipVerify {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA cert: %w", err)
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caPool
	}

	return credentials.NewTLS(tlsConfig), nil
}

func LoadInsecureCredentials() credentials.TransportCredentials {
	return credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
}
