package httpsignatureutils

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadBase64Key loads an Ed25519 private key from a base64-encoded PEM string.
//
// @param base64Key - The base64-encoded PEM private key string.
// @returns The ed25519.PrivateKey or an error if decoding or validation fails.
//
func LoadBase64Key(base64Key string) (ed25519.PrivateKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	if edKey, ok := privateKey.(ed25519.PrivateKey); ok {
		return edKey, nil
	} else {
		return nil, fmt.Errorf("key is not an Ed25519 key")
	}
}

// LoadKey loads an Ed25519 private key from a PEM-encoded file.
//
// @param keyFilePath - The path to the PEM-encoded private key file.
// @returns The ed25519.PrivateKey or an error if loading or validation fails.
//
func LoadKey(keyFilePath string) (ed25519.PrivateKey, error) {
	fileBytes, err := os.ReadFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not load file: %s", keyFilePath)
	}

	block, _ := pem.Decode(fileBytes)
	if block == nil {
		return nil, fmt.Errorf("file was loaded, but did not contain a valid PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("file was loaded, but private key was invalid: %w", err)
	}

	edKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key did not have Ed25519 curve")
	}

	return edKey, nil
}
