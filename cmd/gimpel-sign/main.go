package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/pkg/signing"
)

var rootCmd = &cobra.Command{
	Use:   "gimpel-sign",
	Short: "Module signing and key management for Gimpel",
	Long:  `A CLI tool for managing signing keys and signing module images for the Gimpel honeypot platform.`,
}

var generateKeyCmd = &cobra.Command{
	Use:   "generate-key",
	Short: "Generate a new Ed25519 key pair",
	Long:  `Generate a new Ed25519 key pair for signing modules. The private key should be kept secure and used only on the master server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir, _ := cmd.Flags().GetString("output")
		name, _ := cmd.Flags().GetString("name")

		if outputDir == "" {
			outputDir = "."
		}

		kp, err := signing.GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("generating key pair: %w", err)
		}

		privPath := fmt.Sprintf("%s/%s.key", outputDir, name)
		pubPath := fmt.Sprintf("%s/%s.pub", outputDir, name)

		if err := kp.SavePrivateKey(privPath); err != nil {
			return fmt.Errorf("saving private key: %w", err)
		}

		if err := kp.SavePublicKey(pubPath); err != nil {
			os.Remove(privPath)
			return fmt.Errorf("saving public key: %w", err)
		}

		fmt.Printf("Key pair generated successfully!\n")
		fmt.Printf("  Key ID:      %s\n", kp.KeyID)
		fmt.Printf("  Private key: %s (keep this secure!)\n", privPath)
		fmt.Printf("  Public key:  %s (distribute to agents)\n", pubPath)

		return nil
	},
}

var showKeyCmd = &cobra.Command{
	Use:   "show-key [key-file]",
	Short: "Display information about a key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyFile := args[0]

		kp, err := signing.LoadPrivateKey(keyFile)
		if err != nil {
			kp, err = signing.LoadPublicKey(keyFile)
			if err != nil {
				return fmt.Errorf("loading key: %w", err)
			}
		}

		fmt.Printf("Key Information:\n")
		fmt.Printf("  Key ID:         %s\n", kp.KeyID)
		fmt.Printf("  Has private:    %v\n", kp.PrivateKey != nil)
		fmt.Printf("  Public key len: %d bytes\n", len(kp.PublicKey))

		return nil
	},
}

var signModuleCmd = &cobra.Command{
	Use:   "sign-module",
	Short: "Sign a module image",
	Long:  `Sign a module image file and output the signature. This is typically done during CI/CD or when publishing a new module version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		keyFile, _ := cmd.Flags().GetString("key")
		moduleID, _ := cmd.Flags().GetString("id")
		version, _ := cmd.Flags().GetString("version")
		imageFile, _ := cmd.Flags().GetString("image")
		outputFile, _ := cmd.Flags().GetString("output")

		kp, err := signing.LoadPrivateKey(keyFile)
		if err != nil {
			return fmt.Errorf("loading private key: %w", err)
		}

		imageData, err := os.ReadFile(imageFile)
		if err != nil {
			return fmt.Errorf("reading image file: %w", err)
		}

		digest := signing.ComputeImageDigest(imageData)

		signer, err := signing.NewModuleSigner(kp)
		if err != nil {
			return fmt.Errorf("creating signer: %w", err)
		}

		module := &gimpelv1.ModuleImage{
			Id:        moduleID,
			Version:   version,
			Digest:    digest,
			SizeBytes: int64(len(imageData)),
		}

		if err := signer.SignModule(module); err != nil {
			return fmt.Errorf("signing module: %w", err)
		}

		fmt.Printf("Module signed successfully!\n")
		fmt.Printf("  Module ID: %s\n", moduleID)
		fmt.Printf("  Version:   %s\n", version)
		fmt.Printf("  Digest:    %s\n", digest)
		fmt.Printf("  Signed by: %s\n", module.SignedBy)
		fmt.Printf("  Size:      %d bytes\n", module.SizeBytes)

		if outputFile != "" {
			data := fmt.Sprintf(`{
  "id": "%s",
  "version": "%s",
  "digest": "%s",
  "manifest": "%x",
  "signature": "%x",
  "signed_by": "%s",
  "signed_at": %d,
  "size_bytes": %d
}
`, module.Id, module.Version, module.Digest, module.Manifest,
				module.Signature, module.SignedBy, module.SignedAt, module.SizeBytes)

			if err := os.WriteFile(outputFile, []byte(data), 0644); err != nil {
				return fmt.Errorf("writing output file: %w", err)
			}
			fmt.Printf("  Metadata:  %s\n", outputFile)
		}

		return nil
	},
}

var verifyModuleCmd = &cobra.Command{
	Use:   "verify-module",
	Short: "Verify a module's signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyFile, _ := cmd.Flags().GetString("key")
		moduleID, _ := cmd.Flags().GetString("id")
		version, _ := cmd.Flags().GetString("version")
		imageFile, _ := cmd.Flags().GetString("image")
		manifestHex, _ := cmd.Flags().GetString("manifest")
		signatureHex, _ := cmd.Flags().GetString("signature")

		kp, err := signing.LoadPublicKey(keyFile)
		if err != nil {
			return fmt.Errorf("loading public key: %w", err)
		}

		imageData, err := os.ReadFile(imageFile)
		if err != nil {
			return fmt.Errorf("reading image file: %w", err)
		}
		digest := signing.ComputeImageDigest(imageData)

		var manifest []byte
		if _, err := fmt.Sscanf(manifestHex, "%x", &manifest); err != nil {
			return fmt.Errorf("parsing manifest: %w", err)
		}

		var signature []byte
		if _, err := fmt.Sscanf(signatureHex, "%x", &signature); err != nil {
			return fmt.Errorf("parsing signature: %w", err)
		}

		verifier := signing.NewModuleVerifier(kp)

		module := &gimpelv1.ModuleImage{
			Id:        moduleID,
			Version:   version,
			Digest:    digest,
			Manifest:  manifest,
			Signature: signature,
			SignedBy:  kp.KeyID,
		}

		if err := verifier.VerifyModule(module); err != nil {
			fmt.Printf("❌ Verification FAILED: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Verification PASSED\n")
		fmt.Printf("  Module: %s:%s\n", moduleID, version)
		fmt.Printf("  Digest: %s\n", digest)
		fmt.Printf("  Signed by: %s\n", kp.KeyID)

		return nil
	},
}

func init() {
	generateKeyCmd.Flags().StringP("output", "o", ".", "Output directory for key files")
	generateKeyCmd.Flags().StringP("name", "n", "gimpel", "Base name for key files")

	signModuleCmd.Flags().StringP("key", "k", "", "Path to private key file")
	signModuleCmd.Flags().StringP("id", "i", "", "Module ID")
	signModuleCmd.Flags().StringP("version", "v", "", "Module version")
	signModuleCmd.Flags().String("image", "", "Path to module image file (OCI tarball)")
	signModuleCmd.Flags().StringP("output", "o", "", "Output file for signature metadata (JSON)")
	signModuleCmd.MarkFlagRequired("key")
	signModuleCmd.MarkFlagRequired("id")
	signModuleCmd.MarkFlagRequired("version")
	signModuleCmd.MarkFlagRequired("image")

	verifyModuleCmd.Flags().StringP("key", "k", "", "Path to public key file")
	verifyModuleCmd.Flags().StringP("id", "i", "", "Module ID")
	verifyModuleCmd.Flags().StringP("version", "v", "", "Module version")
	verifyModuleCmd.Flags().String("image", "", "Path to module image file")
	verifyModuleCmd.Flags().String("manifest", "", "Manifest in hex format")
	verifyModuleCmd.Flags().String("signature", "", "Signature in hex format")
	verifyModuleCmd.MarkFlagRequired("key")
	verifyModuleCmd.MarkFlagRequired("id")
	verifyModuleCmd.MarkFlagRequired("version")
	verifyModuleCmd.MarkFlagRequired("image")
	verifyModuleCmd.MarkFlagRequired("manifest")
	verifyModuleCmd.MarkFlagRequired("signature")

	rootCmd.AddCommand(generateKeyCmd)
	rootCmd.AddCommand(showKeyCmd)
	rootCmd.AddCommand(signModuleCmd)
	rootCmd.AddCommand(verifyModuleCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
