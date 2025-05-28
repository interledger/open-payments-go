package httpsignatureutils

import (
	"bytes"
	"crypto/ed25519"
	"net/http"
	"testing"
)

func TestCreateAndValidateSignature(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	body := bytes.NewBufferString("test body")
	req, err := http.NewRequest("POST", "https://example.com/resource", body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	sigHeaders, err := CreateHeaders(SignOptions{
		Request:    req,
		PrivateKey: privateKey,
		KeyID:      "test-key",
	})
	if err != nil {
		t.Fatalf("failed to sign request: %v", err)
	}

	req.Header.Set("Signature", sigHeaders.Signature)
	req.Header.Set("Signature-Input", sigHeaders.SignatureInput)

	err = ValidateSignature(
		NewValidationOptions(req, req.Header, publicKey),
	)
	if err != nil {
		t.Fatalf("signature validation failed: %v", err)
	}

		// negative test - bad signature should fail
		t.Run("InvalidSignature", func(t *testing.T) {
			// flip bits of first byte
			badSig := []byte(sigHeaders.Signature)
			badSig[0] ^= 0xFF
	
			req.Header.Set("Signature", string(badSig))
	
			err = ValidateSignature(
				NewValidationOptions(req, req.Header, publicKey),
			)
			if err == nil {
				t.Fatal("expected validation to fail with bad signature, but it passed")
			}
		})
}
