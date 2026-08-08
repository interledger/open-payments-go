package httpsignatureutils

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"testing"
)

func TestCreateAndValidateSignature(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	bodyBytes := []byte("test body")
	req, err := http.NewRequest("POST", "https://example.com/resource", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	contentHeaders := CreateContentHeaders(bodyBytes)
	req.Header.Set("Content-Digest", contentHeaders.ContentDigest)
	req.Header.Set("Content-Length", contentHeaders.ContentLength)
	req.Header.Set("Content-Type", contentHeaders.ContentType)
	req.ContentLength = int64(len(bodyBytes))

	sigHeaders, err := CreateSignatureHeaders(SignOptions{
		Request:    req,
		PrivateKey: privateKey,
		KeyID:      "test-key",
	})
	if err != nil {
		t.Fatalf("failed to sign request: %v", err)
	}

	req.Header.Set("Signature", sigHeaders.Signature)
	req.Header.Set("Signature-Input", sigHeaders.SignatureInput)
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	err = ValidateSignature(
		NewValidationOptions(req, req.Header, publicKey),
	)
	if err != nil {
		t.Fatalf("signature validation failed: %v", err)
	}

	// negative test - bad signature should fail
	t.Run("InvalidSignature", func(t *testing.T) {
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		sigBytes, err := base64.StdEncoding.DecodeString(sigHeaders.Signature)
		if err != nil {
			t.Fatalf("decode signature: %v", err)
		}
		sigBytes[0] ^= 0xFF
		req.Header.Set("Signature", base64.StdEncoding.EncodeToString(sigBytes))

		err = ValidateSignature(
			NewValidationOptions(req, req.Header, publicKey),
		)
		if err == nil {
			t.Fatal("expected validation to fail with bad signature, but it passed")
		}
	})
}

func TestValidateSignature_RejectsUnderCoveredComponents(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	body := []byte(`{"amount":"1"}`)
	req, err := http.NewRequest("POST", "https://example.com/resource", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))

	// Sign only content-type (missing @method/@target-uri/content-digest).
	created := int64(1700000000)
	keyID := "test-key"
	components := []string{"content-type"}
	base, err := createSignatureBaseString(req, components, created, keyID)
	if err != nil {
		t.Fatalf("base: %v", err)
	}
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, []byte(base)))
	req.Header.Set("Signature", sig)
	req.Header.Set("Signature-Input", `sig1=("content-type");created=1700000000;keyid="test-key";alg="ed25519"`)

	err = ValidateSignature(NewValidationOptions(req, req.Header, publicKey))
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature for under-covered input, got %v", err)
	}
}

func TestValidateSignature_RejectsBodyDigestMismatch(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	original := []byte(`{"amount":"1"}`)
	req, err := http.NewRequest("POST", "https://example.com/resource", bytes.NewReader(original))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	contentHeaders := CreateContentHeaders(original)
	req.Header.Set("Content-Digest", contentHeaders.ContentDigest)
	req.Header.Set("Content-Length", contentHeaders.ContentLength)
	req.Header.Set("Content-Type", contentHeaders.ContentType)
	req.ContentLength = int64(len(original))

	sigHeaders, err := CreateSignatureHeaders(SignOptions{
		Request:    req,
		PrivateKey: privateKey,
		KeyID:      "test-key",
	})
	if err != nil {
		t.Fatalf("failed to sign request: %v", err)
	}
	req.Header.Set("Signature", sigHeaders.Signature)
	req.Header.Set("Signature-Input", sigHeaders.SignatureInput)

	// Swap body under the original Content-Digest + Signature headers.
	tampered := []byte(`{"amount":"999"}`)
	req.Body = io.NopCloser(bytes.NewReader(tampered))
	req.ContentLength = int64(len(tampered))
	req.Header.Set("Content-Length", strconv.Itoa(len(tampered)))

	err = ValidateSignature(NewValidationOptions(req, req.Header, publicKey))
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature for body/digest mismatch, got %v", err)
	}
}
