package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"gimpel/internal/master/config"
)

type CA struct {
	cfg         *config.CAConfig
	certificate *x509.Certificate
	privateKey  *rsa.PrivateKey

	certPEM []byte
}

func New(cfg *config.CAConfig) (*CA, error) {
	ca := &CA{cfg: cfg}

	if err := ca.load(); err != nil {
		if !cfg.AutoGenerate {
			return nil, fmt.Errorf("loading CA: %w", err)
		}
		if err := ca.generate(); err != nil {
			return nil, fmt.Errorf("generating CA: %w", err)
		}
	}

	return ca, nil
}

func (ca *CA) load() error {
	certPEM, err := os.ReadFile(ca.cfg.CertFile)
	if err != nil {
		return fmt.Errorf("reading cert: %w", err)
	}

	keyPEM, err := os.ReadFile(ca.cfg.KeyFile)
	if err != nil {
		return fmt.Errorf("reading key: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to decode cert PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parsing cert: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode key PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parsing key: %w", err)
	}

	ca.certificate = cert
	ca.privateKey = key
	ca.certPEM = certPEM
	return nil
}

func (ca *CA) generate() error {
	key, err := rsa.GenerateKey(rand.Reader, ca.cfg.KeySize)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generating serial: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{ca.cfg.Organization},
			CommonName:   ca.cfg.Organization + " CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, ca.cfg.ValidityDays),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("creating certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return fmt.Errorf("parsing created cert: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(ca.cfg.CertFile), 0700); err != nil {
		return fmt.Errorf("creating dir: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(ca.cfg.CertFile, certPEM, 0644); err != nil {
		return fmt.Errorf("writing cert: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(ca.cfg.KeyFile, keyPEM, 0600); err != nil {
		return fmt.Errorf("writing key: %w", err)
	}

	ca.certificate = cert
	ca.privateKey = key
	ca.certPEM = certPEM
	return nil
}

func (ca *CA) CACertPEM() []byte {
	return ca.certPEM
}

type CertRequest struct {
	AgentID   string
	Hostname  string
	PublicIPs []string
}

type SignedCert struct {
	Certificate []byte
	PrivateKey  []byte
}

func (ca *CA) IssueCertificate(req *CertRequest) (*SignedCert, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("generating serial: %w", err)
	}

	ips := make([]net.IP, 0, len(req.PublicIPs))
	for _, ipStr := range req.PublicIPs {
		if ip := net.ParseIP(ipStr); ip != nil {
			ips = append(ips, ip)
		}
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{ca.cfg.Organization},
			CommonName:   req.AgentID,
		},
		DNSNames:    []string{req.Hostname},
		IPAddresses: ips,
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, ca.cfg.ValidityDays),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, ca.certificate, &key.PublicKey, ca.privateKey)
	if err != nil {
		return nil, fmt.Errorf("signing certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return &SignedCert{
		Certificate: certPEM,
		PrivateKey:  keyPEM,
	}, nil
}
