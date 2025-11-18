package httpsignatureutils

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to generate a test Ed25519 private key in PEM format
func generateTestKeyPEM(t *testing.T) (ed25519.PrivateKey, []byte) {
	t.Helper()
	
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Marshal to PKCS8
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal key: %v", err)
	}

	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	pemBytes := pem.EncodeToMemory(pemBlock)
	return privateKey, pemBytes
}

func TestLoadBase64Key(t *testing.T) {
	t.Run("valid base64 encoded key", func(t *testing.T) {
		originalKey, pemBytes := generateTestKeyPEM(t)
		base64Key := base64.StdEncoding.EncodeToString(pemBytes)

		loadedKey, err := LoadBase64Key(base64Key)
		if err != nil {
			t.Fatalf("LoadBase64Key failed: %v", err)
		}

		if !loadedKey.Equal(originalKey) {
			t.Error("loaded key does not match original key")
		}
	})

	t.Run("invalid base64 string", func(t *testing.T) {
		_, err := LoadBase64Key("this is not valid base64!!!")
		if err == nil {
			t.Error("expected error for invalid base64, got nil")
		}
		if err != nil && !strings.HasPrefix(err.Error(), "invalid base64:") {
			t.Errorf("expected error to start with 'invalid base64:', got: %v", err)
		}
	})

	t.Run("valid base64 but not PEM", func(t *testing.T) {
		notPEM := base64.StdEncoding.EncodeToString([]byte("not a PEM block"))
		_, err := LoadBase64Key(notPEM)
		if err == nil {
			t.Error("expected error for non-PEM data, got nil")
		}
		if err != nil && err.Error() != "failed to parse PEM block" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("valid PEM but invalid private key", func(t *testing.T) {
		invalidPEM := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: []byte("invalid key data"),
		}
		pemBytes := pem.EncodeToMemory(invalidPEM)
		base64Key := base64.StdEncoding.EncodeToString(pemBytes)

		_, err := LoadBase64Key(base64Key)
		if err == nil {
			t.Error("expected error for invalid private key, got nil")
		}
	})
}

func TestLoadKey(t *testing.T) {
	t.Run("valid key file", func(t *testing.T) {
		// Create temporary key file
		originalKey, pemBytes := generateTestKeyPEM(t)
		
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "test_key.pem")
		
		err := os.WriteFile(keyPath, pemBytes, 0600)
		if err != nil {
			t.Fatalf("failed to write test key file: %v", err)
		}

		loadedKey, err := LoadKey(keyPath)
		if err != nil {
			t.Fatalf("LoadKey failed: %v", err)
		}

		if !loadedKey.Equal(originalKey) {
			t.Error("loaded key does not match original key")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		_, err := LoadKey("/nonexistent/path/to/key.pem")
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
		if err != nil && err.Error() != "could not load file: /nonexistent/path/to/key.pem" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("file exists but not PEM", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "not_pem.txt")
		
		err := os.WriteFile(keyPath, []byte("not a PEM file"), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err = LoadKey(keyPath)
		if err == nil {
			t.Error("expected error for non-PEM file, got nil")
		}
		if err != nil && err.Error() != "file was loaded, but did not contain a valid PEM block" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("valid PEM but invalid private key", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "invalid_key.pem")
		
		invalidPEM := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: []byte("invalid key data"),
		}
		pemBytes := pem.EncodeToMemory(invalidPEM)
		
		err := os.WriteFile(keyPath, pemBytes, 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err = LoadKey(keyPath)
		if err == nil {
			t.Error("expected error for invalid private key, got nil")
		}
	})

	t.Run("valid key but not Ed25519", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "wrong_type.pem")
		
		// Create invalid PKCS8 that will parse but not be Ed25519
		invalidPEM := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: []byte{0x30, 0x82, 0x01, 0x00},
		}
		pemBytes := pem.EncodeToMemory(invalidPEM)
		
		err := os.WriteFile(keyPath, pemBytes, 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err = LoadKey(keyPath)
		if err == nil {
			t.Error("expected error for non-Ed25519 key, got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "empty.pem")
		
		err := os.WriteFile(keyPath, []byte{}, 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err = LoadKey(keyPath)
		if err == nil {
			t.Error("expected error for empty file, got nil")
		}
		if err != nil && err.Error() != "file was loaded, but did not contain a valid PEM block" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestLoadBase64KeyAndLoadKeyEquivalence(t *testing.T) {
	t.Run("both functions load same key", func(t *testing.T) {
		// Generate test key
		_, pemBytes := generateTestKeyPEM(t)
		
		// Test LoadBase64Key
		base64Key := base64.StdEncoding.EncodeToString(pemBytes)
		key1, err := LoadBase64Key(base64Key)
		if err != nil {
			t.Fatalf("LoadBase64Key failed: %v", err)
		}

		// Test LoadKey
		tmpDir := t.TempDir()
		keyPath := filepath.Join(tmpDir, "test_key.pem")
		err = os.WriteFile(keyPath, pemBytes, 0600)
		if err != nil {
			t.Fatalf("failed to write test key file: %v", err)
		}

		key2, err := LoadKey(keyPath)
		if err != nil {
			t.Fatalf("LoadKey failed: %v", err)
		}

		// Compare keys
		if !key1.Equal(key2) {
			t.Error("LoadBase64Key and LoadKey returned different keys for same PEM data")
		}
	})
}