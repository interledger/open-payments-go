package httpsignatureutils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCreateSignatureHeaders_Basic(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	req, _ := http.NewRequest("GET", "https://example.com/api/resource", nil)
	req.Header.Set("Content-Type", "application/json")

	opts := SignOptions{
		Request:    req,
		PrivateKey: priv,
		KeyID:      "test-key-1",
	}

	headers, err := CreateSignatureHeaders(opts)
	if err != nil {
		t.Fatalf("CreateSignatureHeaders returned error: %v", err)
	}

	if headers.Signature == "" {
		t.Error("Expected non-empty Signature")
	}
	if headers.SignatureInput == "" {
		t.Error("Expected non-empty SignatureInput")
	}

	for _, comp := range []string{"@method", "@target-uri"} {
		if !strings.Contains(headers.SignatureInput, comp) {
			t.Errorf("SignatureInput missing component: %s", comp)
		}
	}

	if !strings.Contains(headers.SignatureInput, "keyid=\"test-key-1\"") {
		t.Error("SignatureInput missing correct keyid")
	}

	// verify signature
	created := time.Now().Unix()
	baseString, err := createSignatureBaseString(req, []string{"@method", "@target-uri"}, created, opts.KeyID)
	if err != nil {
		t.Fatalf("Failed to create signature base string: %v", err)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(headers.Signature)
	if err != nil {
		t.Fatalf("Failed to decode signature: %v", err)
	}
	if !ed25519.Verify(pub, []byte(baseString), sigBytes) {
		t.Error("Signature verification failed")
	}
}

func TestCreateSignatureHeaders_WithAuthorization(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	req, _ := http.NewRequest("POST", "https://example.com/api/resource", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	opts := SignOptions{
		Request:    req,
		PrivateKey: priv,
		KeyID:      "test-key-2",
	}

	headers, err := CreateSignatureHeaders(opts)
	if err != nil {
		t.Fatalf("CreateSignatureHeaders returned error: %v", err)
	}

	if !strings.Contains(headers.SignatureInput, "authorization") {
		t.Error("SignatureInput should contain 'authorization' when header is present")
	}
}

func TestCreateSignatureHeaders_WithBody(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	body := strings.NewReader(`{"field":"value"}`)
	req, _ := http.NewRequest("POST", "https://example.com/api/resource", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Digest", "sha-256=xyz")
	req.Header.Set("Content-Length", "15")

	opts := SignOptions{
		Request:    req,
		PrivateKey: priv,
		KeyID:      "test-key-3",
	}

	headers, err := CreateSignatureHeaders(opts)
	if err != nil {
		t.Fatalf("CreateSignatureHeaders returned error: %v", err)
	}

	if !strings.Contains(headers.SignatureInput, "content-digest") {
		t.Error("SignatureInput should contain 'content-digest' when body is present")
	}
	if !strings.Contains(headers.SignatureInput, "content-length") {
		t.Error("SignatureInput should contain 'content-length' when body is present")
	}
}
