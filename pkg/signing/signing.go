// Package signing provides Ed25519 signature utilities for module verification.
package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

const (
	PrivateKeyPEMType = "GIMPEL PRIVATE KEY"
	PublicKeyPEMType  = "GIMPEL PUBLIC KEY"
)

type KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	KeyID      string
}

func GenerateKeyPair() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating key pair: %w", err)
	}

	keyID := computeKeyID(pub)

	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
		KeyID:      keyID,
	}, nil
}

func computeKeyID(pub ed25519.PublicKey) string {
	hash := sha256.Sum256(pub)
	return hex.EncodeToString(hash[:8])
}

func (kp *KeyPair) Sign(data []byte) []byte {
	return ed25519.Sign(kp.PrivateKey, data)
}

func (kp *KeyPair) Verify(data, signature []byte) bool {
	return ed25519.Verify(kp.PublicKey, data, signature)
}

func (kp *KeyPair) SavePrivateKey(path string) error {
	block := &pem.Block{
		Type:  PrivateKeyPEMType,
		Bytes: kp.PrivateKey,
		Headers: map[string]string{
			"Key-ID":     kp.KeyID,
			"Created-At": time.Now().UTC().Format(time.RFC3339),
		},
	}

	data := pem.EncodeToMemory(block)
	return os.WriteFile(path, data, 0600)
}

func (kp *KeyPair) SavePublicKey(path string) error {
	block := &pem.Block{
		Type:  PublicKeyPEMType,
		Bytes: kp.PublicKey,
		Headers: map[string]string{
			"Key-ID": kp.KeyID,
		},
	}

	data := pem.EncodeToMemory(block)
	return os.WriteFile(path, data, 0644)
}

func LoadPrivateKey(path string) (*KeyPair, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in file")
	}

	if block.Type != PrivateKeyPEMType {
		return nil, fmt.Errorf("unexpected PEM type: %s", block.Type)
	}

	if len(block.Bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: %d", len(block.Bytes))
	}

	priv := ed25519.PrivateKey(block.Bytes)
	pub := priv.Public().(ed25519.PublicKey)
	keyID := computeKeyID(pub)

	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: priv,
		KeyID:      keyID,
	}, nil
}

func LoadPublicKey(path string) (*KeyPair, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in file")
	}

	if block.Type != PublicKeyPEMType {
		return nil, fmt.Errorf("unexpected PEM type: %s", block.Type)
	}

	if len(block.Bytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(block.Bytes))
	}

	pub := ed25519.PublicKey(block.Bytes)
	keyID := computeKeyID(pub)

	return &KeyPair{
		PublicKey:  pub,
		PrivateKey: nil,
		KeyID:      keyID,
	}, nil
}

type Verifier struct {
	trustedKeys map[string]ed25519.PublicKey
}

func NewVerifier(keys ...*KeyPair) *Verifier {
	v := &Verifier{
		trustedKeys: make(map[string]ed25519.PublicKey),
	}
	for _, kp := range keys {
		v.AddTrustedKey(kp.KeyID, kp.PublicKey)
	}
	return v
}

func (v *Verifier) AddTrustedKey(keyID string, pub ed25519.PublicKey) {
	v.trustedKeys[keyID] = pub
}

func (v *Verifier) Verify(data, signature []byte, keyID string) error {
	pub, ok := v.trustedKeys[keyID]
	if !ok {
		return fmt.Errorf("unknown key ID: %s", keyID)
	}

	if !ed25519.Verify(pub, data, signature) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

func (v *Verifier) HasTrustedKey(keyID string) bool {
	_, ok := v.trustedKeys[keyID]
	return ok
}

func (v *Verifier) TrustedKeyIDs() []string {
	ids := make([]string, 0, len(v.trustedKeys))
	for id := range v.trustedKeys {
		ids = append(ids, id)
	}
	return ids
}
