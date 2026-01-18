package signing

import (
	"os"
	"path/filepath"
	"testing"

	gimpelv1 "gimpel/api/go/v1"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	if kp.PublicKey == nil {
		t.Error("PublicKey is nil")
	}
	if kp.PrivateKey == nil {
		t.Error("PrivateKey is nil")
	}
	if kp.KeyID == "" {
		t.Error("KeyID is empty")
	}
	if len(kp.KeyID) != 16 {
		t.Errorf("KeyID length = %d, want 16", len(kp.KeyID))
	}
}

func TestSignAndVerify(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	data := []byte("test data to sign")
	signature := kp.Sign(data)

	if !kp.Verify(data, signature) {
		t.Error("Verify returned false for valid signature")
	}

	modifiedData := []byte("modified data")
	if kp.Verify(modifiedData, signature) {
		t.Error("Verify returned true for modified data")
	}

	modifiedSig := make([]byte, len(signature))
	copy(modifiedSig, signature)
	modifiedSig[0] ^= 0xFF
	if kp.Verify(data, modifiedSig) {
		t.Error("Verify returned true for modified signature")
	}
}

func TestSaveAndLoadKeys(t *testing.T) {
	tmpDir := t.TempDir()

	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	privPath := filepath.Join(tmpDir, "test.key")
	pubPath := filepath.Join(tmpDir, "test.pub")

	if err := kp.SavePrivateKey(privPath); err != nil {
		t.Fatalf("SavePrivateKey failed: %v", err)
	}
	if err := kp.SavePublicKey(pubPath); err != nil {
		t.Fatalf("SavePublicKey failed: %v", err)
	}

	loadedPriv, err := LoadPrivateKey(privPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}
	if loadedPriv.KeyID != kp.KeyID {
		t.Errorf("LoadPrivateKey KeyID = %s, want %s", loadedPriv.KeyID, kp.KeyID)
	}
	if loadedPriv.PrivateKey == nil {
		t.Error("LoadPrivateKey: PrivateKey is nil")
	}

	loadedPub, err := LoadPublicKey(pubPath)
	if err != nil {
		t.Fatalf("LoadPublicKey failed: %v", err)
	}
	if loadedPub.KeyID != kp.KeyID {
		t.Errorf("LoadPublicKey KeyID = %s, want %s", loadedPub.KeyID, kp.KeyID)
	}
	if loadedPub.PrivateKey != nil {
		t.Error("LoadPublicKey: PrivateKey should be nil")
	}

	data := []byte("cross-sign test")
	sig := kp.Sign(data)
	if !loadedPub.Verify(data, sig) {
		t.Error("Cross verification failed")
	}
}

func TestPrivateKeyPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	kp, _ := GenerateKeyPair()
	privPath := filepath.Join(tmpDir, "test.key")

	if err := kp.SavePrivateKey(privPath); err != nil {
		t.Fatalf("SavePrivateKey failed: %v", err)
	}

	info, err := os.Stat(privPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Private key permissions = %o, want 0600", mode)
	}
}

func TestVerifier(t *testing.T) {
	kp1, _ := GenerateKeyPair()
	kp2, _ := GenerateKeyPair()

	verifier := NewVerifier(kp1)

	data := []byte("test data")
	sig1 := kp1.Sign(data)
	sig2 := kp2.Sign(data)

	if err := verifier.Verify(data, sig1, kp1.KeyID); err != nil {
		t.Errorf("Verify failed for trusted key: %v", err)
	}

	if err := verifier.Verify(data, sig2, kp2.KeyID); err == nil {
		t.Error("Verify should fail for untrusted key")
	}

	verifier.AddTrustedKey(kp2.KeyID, kp2.PublicKey)
	if err := verifier.Verify(data, sig2, kp2.KeyID); err != nil {
		t.Errorf("Verify failed after adding trusted key: %v", err)
	}
}

func TestModuleSigning(t *testing.T) {
	kp, _ := GenerateKeyPair()
	signer, err := NewModuleSigner(kp)
	if err != nil {
		t.Fatalf("NewModuleSigner failed: %v", err)
	}

	module := &gimpelv1.ModuleImage{
		Id:      "test-module",
		Version: "1.0.0",
		Digest:  "sha256:abc123",
	}

	if err := signer.SignModule(module); err != nil {
		t.Fatalf("SignModule failed: %v", err)
	}

	if module.Signature == nil {
		t.Error("Module signature is nil after signing")
	}
	if module.SignedBy != kp.KeyID {
		t.Errorf("SignedBy = %s, want %s", module.SignedBy, kp.KeyID)
	}
	if module.SignedAt == 0 {
		t.Error("SignedAt is 0")
	}

	verifier := NewModuleVerifier(kp)
	if err := verifier.VerifyModule(module); err != nil {
		t.Errorf("VerifyModule failed: %v", err)
	}

	module.Id = "tampered-module"
	if err := verifier.VerifyModule(module); err == nil {
		t.Error("VerifyModule should fail for tampered module")
	}
}

func TestCatalogSigning(t *testing.T) {
	kp, _ := GenerateKeyPair()
	signer, _ := NewModuleSigner(kp)

	catalog := &gimpelv1.ModuleCatalog{
		Version: 1,
		Modules: []*gimpelv1.ModuleImage{
			{Id: "module1", Version: "1.0.0", Digest: "sha256:aaa"},
			{Id: "module2", Version: "2.0.0", Digest: "sha256:bbb"},
		},
	}

	if err := signer.SignCatalog(catalog); err != nil {
		t.Fatalf("SignCatalog failed: %v", err)
	}

	if catalog.Signature == nil {
		t.Error("Catalog signature is nil")
	}

	verifier := NewModuleVerifier(kp)
	if err := verifier.VerifyCatalog(catalog); err != nil {
		t.Errorf("VerifyCatalog failed: %v", err)
	}

	catalog.Version = 999
	if err := verifier.VerifyCatalog(catalog); err == nil {
		t.Error("VerifyCatalog should fail for tampered catalog")
	}
}

func TestComputeImageDigest(t *testing.T) {
	data := []byte("test image data")
	digest := ComputeImageDigest(data)

	if digest == "" {
		t.Error("Digest is empty")
	}
	if len(digest) != 71 {
		t.Errorf("Digest length = %d, want 71", len(digest))
	}
	if digest[:7] != "sha256:" {
		t.Errorf("Digest prefix = %s, want sha256:", digest[:7])
	}

	digest2 := ComputeImageDigest(data)
	if digest != digest2 {
		t.Error("Same data produced different digests")
	}

	digest3 := ComputeImageDigest([]byte("different data"))
	if digest == digest3 {
		t.Error("Different data produced same digest")
	}
}

func TestLoadKeyErrors(t *testing.T) {
	_, err := LoadPrivateKey("/nonexistent/path")
	if err == nil {
		t.Error("LoadPrivateKey should fail for non-existent file")
	}

	_, err = LoadPublicKey("/nonexistent/path")
	if err == nil {
		t.Error("LoadPublicKey should fail for non-existent file")
	}

	tmpDir := t.TempDir()
	badPath := filepath.Join(tmpDir, "bad.key")
	if err := os.WriteFile(badPath, []byte("not valid PEM"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	_, err = LoadPrivateKey(badPath)
	if err == nil {
		t.Error("LoadPrivateKey should fail for invalid PEM")
	}

	_, err = LoadPublicKey(badPath)
	if err == nil {
		t.Error("LoadPublicKey should fail for invalid PEM")
	}
}

func TestVerifyWithWrongKeyID(t *testing.T) {
	kp, _ := GenerateKeyPair()
	verifier := NewVerifier(kp)

	data := []byte("test")
	sig := kp.Sign(data)

	err := verifier.Verify(data, sig, "wrong-key-id")
	if err == nil {
		t.Error("Verify should fail with wrong key ID")
	}
}

func TestSignerRequiresPrivateKey(t *testing.T) {
	kp, _ := GenerateKeyPair()

	tmpDir := t.TempDir()
	pubPath := filepath.Join(tmpDir, "test.pub")
	kp.SavePublicKey(pubPath)
	pubOnly, _ := LoadPublicKey(pubPath)

	_, err := NewModuleSigner(pubOnly)
	if err == nil {
		t.Error("NewModuleSigner should fail without private key")
	}
}

func TestAgentConfigSigning(t *testing.T) {
	kp, _ := GenerateKeyPair()
	signer, _ := NewModuleSigner(kp)
	verifier := NewModuleVerifier(kp)

	config := &gimpelv1.AgentModuleConfig{
		AgentId: "agent-001",
		Assignments: []*gimpelv1.ModuleAssignment{
			{
				ModuleId: "ssh-honeypot",
				Version:  "1.0.0",
			},
		},
	}

	if err := signer.SignAgentConfig(config); err != nil {
		t.Fatalf("SignAgentConfig failed: %v", err)
	}

	if config.Signature == nil {
		t.Error("Config signature is nil")
	}

	if err := verifier.VerifyAgentConfig(config); err != nil {
		t.Errorf("VerifyAgentConfig failed: %v", err)
	}

	config.AgentId = "tampered-agent"
	if err := verifier.VerifyAgentConfig(config); err == nil {
		t.Error("VerifyAgentConfig should fail for tampered config")
	}
}

func TestEmptyCatalog(t *testing.T) {
	kp, _ := GenerateKeyPair()
	signer, _ := NewModuleSigner(kp)
	verifier := NewModuleVerifier(kp)

	catalog := &gimpelv1.ModuleCatalog{
		Version: 1,
		Modules: []*gimpelv1.ModuleImage{},
	}

	if err := signer.SignCatalog(catalog); err != nil {
		t.Fatalf("SignCatalog failed for empty catalog: %v", err)
	}

	if err := verifier.VerifyCatalog(catalog); err != nil {
		t.Errorf("VerifyCatalog failed for empty catalog: %v", err)
	}
}

func TestKeyIDDeterministic(t *testing.T) {
	kp1, _ := GenerateKeyPair()
	kp2, _ := GenerateKeyPair()

	if kp1.KeyID == kp2.KeyID {
		t.Error("Different key pairs should have different KeyIDs")
	}

	tmpDir := t.TempDir()
	privPath := filepath.Join(tmpDir, "test.key")
	kp1.SavePrivateKey(privPath)

	loaded, _ := LoadPrivateKey(privPath)
	if loaded.KeyID != kp1.KeyID {
		t.Errorf("Reloaded KeyID = %s, want %s", loaded.KeyID, kp1.KeyID)
	}
}
