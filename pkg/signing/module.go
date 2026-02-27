package signing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	gimpelv1 "gimpel/api/go/v1"

	"google.golang.org/protobuf/proto"
)

type ModuleSigner struct {
	keyPair *KeyPair
}

func NewModuleSigner(kp *KeyPair) (*ModuleSigner, error) {
	if kp.PrivateKey == nil {
		return nil, fmt.Errorf("private key required for signing")
	}
	return &ModuleSigner{keyPair: kp}, nil
}

func (s *ModuleSigner) SignModule(module *gimpelv1.ModuleImage) error {
	manifest := &gimpelv1.ModuleManifest{
		ModuleId:      module.Id,
		Version:       module.Version,
		PayloadSha256: module.Digest,
		Timestamp:     time.Now().Unix(),
	}

	manifestBytes, err := proto.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	hash := sha256.Sum256(manifestBytes)
	signature := s.keyPair.Sign(hash[:])

	module.Manifest = manifestBytes
	module.Signature = signature
	module.SignedBy = s.keyPair.KeyID
	module.SignedAt = manifest.Timestamp

	return nil
}

func (s *ModuleSigner) SignCatalog(catalog *gimpelv1.ModuleCatalog) error {
	catalog.Signature = nil
	catalog.SignedBy = ""

	data, err := proto.Marshal(catalog)
	if err != nil {
		return fmt.Errorf("marshaling catalog: %w", err)
	}

	hash := sha256.Sum256(data)
	signature := s.keyPair.Sign(hash[:])

	catalog.Signature = signature
	catalog.SignedBy = s.keyPair.KeyID

	return nil
}

func (s *ModuleSigner) SignAgentConfig(config *gimpelv1.AgentModuleConfig) error {
	config.Signature = nil

	data, err := proto.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling agent config: %w", err)
	}

	hash := sha256.Sum256(data)
	config.Signature = s.keyPair.Sign(hash[:])

	return nil
}

func (s *ModuleSigner) KeyID() string {
	return s.keyPair.KeyID
}

type ModuleVerifier struct {
	verifier *Verifier
}

func NewModuleVerifier(trustedKeys ...*KeyPair) *ModuleVerifier {
	return &ModuleVerifier{
		verifier: NewVerifier(trustedKeys...),
	}
}

func (v *ModuleVerifier) AddTrustedKey(kp *KeyPair) {
	v.verifier.AddTrustedKey(kp.KeyID, kp.PublicKey)
}

func (v *ModuleVerifier) VerifyModule(module *gimpelv1.ModuleImage) error {
	if module.Signature == nil {
		return fmt.Errorf("module is not signed")
	}

	if module.Manifest == nil {
		return fmt.Errorf("module manifest is missing")
	}

	if module.SignedBy == "" {
		return fmt.Errorf("module has no signer key ID")
	}

	var manifest gimpelv1.ModuleManifest
	if err := proto.Unmarshal(module.Manifest, &manifest); err != nil {
		return fmt.Errorf("unmarshaling manifest: %w", err)
	}

	if manifest.ModuleId != module.Id {
		return fmt.Errorf("manifest module ID mismatch: expected %s, got %s", module.Id, manifest.ModuleId)
	}

	if manifest.PayloadSha256 != module.Digest {
		return fmt.Errorf("manifest payload hash mismatch: expected %s, got %s", module.Digest, manifest.PayloadSha256)
	}

	hash := sha256.Sum256(module.Manifest)

	if err := v.verifier.Verify(hash[:], module.Signature, module.SignedBy); err != nil {
		return fmt.Errorf("module signature verification failed: %w", err)
	}

	return nil
}

func (v *ModuleVerifier) VerifyCatalog(catalog *gimpelv1.ModuleCatalog) error {
	if catalog.Signature == nil {
		return fmt.Errorf("catalog is not signed")
	}

	if catalog.SignedBy == "" {
		return fmt.Errorf("catalog has no signer key ID")
	}

	signature := catalog.Signature
	signedBy := catalog.SignedBy
	catalog.Signature = nil
	catalog.SignedBy = ""

	data, err := proto.Marshal(catalog)
	if err != nil {
		catalog.Signature = signature
		catalog.SignedBy = signedBy
		return fmt.Errorf("marshaling catalog: %w", err)
	}

	catalog.Signature = signature
	catalog.SignedBy = signedBy

	hash := sha256.Sum256(data)
	if err := v.verifier.Verify(hash[:], signature, signedBy); err != nil {
		return fmt.Errorf("catalog signature verification failed: %w", err)
	}

	return nil
}

func (v *ModuleVerifier) VerifyAgentConfig(config *gimpelv1.AgentModuleConfig) error {
	if config.Signature == nil {
		return fmt.Errorf("config is not signed")
	}

	signature := config.Signature
	config.Signature = nil

	data, err := proto.Marshal(config)
	if err != nil {
		config.Signature = signature
		return fmt.Errorf("marshaling config: %w", err)
	}

	config.Signature = signature

	hash := sha256.Sum256(data)
	for _, keyID := range v.verifier.TrustedKeyIDs() {
		if err := v.verifier.Verify(hash[:], signature, keyID); err == nil {
			return nil
		}
	}

	return fmt.Errorf("config signature verification failed: no trusted key matched")
}

func ComputeImageDigest(data []byte) string {
	hash := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(hash[:])
}
