package httpsignatureutils

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
)

func LoadBase64Key(base64Key string) (ed25519.PrivateKey, error) {
	// base64 key in bruno decoded should be:
	// -----BEGIN PRIVATE KEY-----
	// MC4CAQAwBQYDK2VwBCIEIEqezmcPhOE8bkwN+jQrppfRYzGIdFTVWQGTHJIKpz88
	// -----END PRIVATE KEY-----
	

	pemBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	// Step : Decode the PEM block
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	
	blockJSON, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JWK:", err)
		return nil, fmt.Errorf("could not marshal json")
	}
	fmt.Println("Parsed key from PEM file:", string(blockJSON))

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	if edKey, ok := privateKey.(ed25519.PrivateKey); ok {
		fmt.Printf("Ed25519 Private Key: %x\n", edKey)
		return edKey, nil
	} else {
		return nil, fmt.Errorf("key is not an Ed25519 key")
	}
}
