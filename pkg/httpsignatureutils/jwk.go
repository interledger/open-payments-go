// TODO ts client has key and jwk files. everything ehre is unused so far but dumping for preservation
// purposes. need to split these out if i actually end up using them.
package httpsignatureutils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type JWK struct {
	Kid string `json:"kid"`
	X   string `json:"x"`
	Alg string `json:"alg"`
	Kty string `json:"kty"`
	Crv string `json:"crv"`
}

func GenerateNewPrivateKey() (string, error) {
	// Generate a new Ed25519 key pair
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", err
	}

	encodedKey := base64.StdEncoding.EncodeToString(privateKey)

	return encodedKey, nil
}

// TODO: rm if unused
func generateJSONWebKey(){
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Println("Error generating key:", err)
		return
	}

	x := base64.RawURLEncoding.EncodeToString(publicKey)

	jwk := JWK{
		Kid: "keyid-123", // Replace with your key ID
		X:   x,
		Alg: "EdDSA",
		Kty: "OKP",
		Crv: "Ed25519",
	}

	jwkJSON, err := json.MarshalIndent(jwk, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JWK:", err)
		return
	}

	fmt.Println(string(jwkJSON))

	privateKeyBase64 := base64.RawURLEncoding.EncodeToString(privateKey.Seed())
	fmt.Println("Private Key (base64):", privateKeyBase64)
}