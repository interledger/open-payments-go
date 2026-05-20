package httpsignatureutils

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

func parseEd25519PEM(pemBytes []byte) (ed25519.PrivateKey, error) {
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

func LoadBase64Key(base64Key string) (ed25519.PrivateKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	return parseEd25519PEM(pemBytes)
}

func LoadKeyFromFile(path string) (ed25519.PrivateKey, error) {
	fileBytes, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("could not load file: %w", err)
	}

	return parseEd25519PEM(fileBytes)
}

func LoadPEMKey(pemString string) (ed25519.PrivateKey, error) {
	return parseEd25519PEM([]byte(pemString))
}

func LoadKey(input string) (ed25519.PrivateKey, error) {
	// PEM file path
	if _, err := os.Stat(input); err == nil {
		return LoadKeyFromFile(input)
	}

	// Raw PEM string
	if strings.HasPrefix(strings.TrimSpace(input), "-----BEGIN") {
		return LoadPEMKey(input)
	}

	// Base64 Key
	return LoadBase64Key(input)
}
