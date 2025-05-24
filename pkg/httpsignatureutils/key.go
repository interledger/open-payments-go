package httpsignatureutils

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

func LoadBase64Key(base64Key string) (ed25519.PrivateKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	// Step : Decode the PEM block
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
