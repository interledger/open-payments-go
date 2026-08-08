package httpsignatureutils

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrInvalidSignature      = errors.New("invalid signature")
	ErrMissingSignature      = errors.New("missing signature")
	ErrMissingSignatureInput = errors.New("missing signature input")
)

type ValidationOptions struct {
	Request   *http.Request
	Headers   http.Header
	PublicKey ed25519.PublicKey
}

func NewValidationOptions(r *http.Request, headers http.Header, publicKey ed25519.PublicKey) *ValidationOptions {
	return &ValidationOptions{
		Request:   r,
		Headers:   headers,
		PublicKey: publicKey,
	}
}

func parseSignatureInput(input string) ([]string, int64, string, error) {
	input = strings.TrimPrefix(input, "sig1=")

	var components []string
	var created int64
	var keyID string

	for _, part := range strings.Split(input, ";") {
		part = strings.TrimSpace(part)

		switch {
		case strings.HasPrefix(part, "(") && strings.HasSuffix(part, ")"):
			inner := part[1 : len(part)-1]
			parsedComponents := strings.Fields(inner)
			// strip quotes from each parsed component
			for _, comp := range parsedComponents {
				components = append(components, strings.Trim(comp, `"`))
			}
		case strings.HasPrefix(part, "created="):
			val := strings.TrimPrefix(part, "created=")
			t, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, 0, "", ErrInvalidSignature
			}
			created = t

		case strings.HasPrefix(part, "keyid="):
			val := strings.TrimPrefix(part, "keyid=")
			keyID = strings.Trim(val, `"`)
		}
	}

	if len(components) == 0 || created == 0 || keyID == "" {
		return nil, 0, "", ErrInvalidSignature
	}

	return components, created, keyID, nil
}

func componentSet(components []string) map[string]struct{} {
	out := make(map[string]struct{}, len(components))
	for _, c := range components {
		out[strings.ToLower(c)] = struct{}{}
	}
	return out
}

// validateRequiredCoveredComponents enforces Open Payments / GNAP MUST covered
// components before cryptographic verification (parity with open-payments-node).
func validateRequiredCoveredComponents(req *http.Request, components []string) error {
	covered := componentSet(components)
	if _, ok := covered["@method"]; !ok {
		return ErrInvalidSignature
	}
	if _, ok := covered["@target-uri"]; !ok {
		return ErrInvalidSignature
	}
	if req.Header.Get("Authorization") != "" {
		if _, ok := covered["authorization"]; !ok {
			return ErrInvalidSignature
		}
	}
	if requestHasBody(req) {
		if _, ok := covered["content-digest"]; !ok {
			return ErrInvalidSignature
		}
	}
	return nil
}

func requestHasBody(req *http.Request) bool {
	if req.ContentLength > 0 {
		return true
	}
	if cl := req.Header.Get("Content-Length"); cl != "" && cl != "0" {
		return true
	}
	return false
}

func verifyContentDigest(req *http.Request, headers http.Header) error {
	digestHeader := headers.Get("Content-Digest")
	if digestHeader == "" {
		return ErrInvalidSignature
	}
	if req.Body == nil || req.Body == http.NoBody {
		return ErrInvalidSignature
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return ErrInvalidSignature
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	expected := createContentDigest(body)
	if !strings.EqualFold(expected, digestHeader) {
		return ErrInvalidSignature
	}
	return nil
}

func ValidateSignature(opts *ValidationOptions) error {
	sigInput := opts.Headers.Get("Signature-Input")
	if sigInput == "" {
		return ErrMissingSignatureInput
	}

	components, created, keyID, err := parseSignatureInput(sigInput)
	if err != nil {
		return err
	}

	if err := validateRequiredCoveredComponents(opts.Request, components); err != nil {
		return err
	}

	covered := componentSet(components)
	if _, ok := covered["content-digest"]; ok {
		if err := verifyContentDigest(opts.Request, opts.Headers); err != nil {
			return err
		}
	}

	sig := opts.Headers.Get("Signature")
	if sig == "" {
		return ErrMissingSignature
	}

	baseString, err := createSignatureBaseString(opts.Request, components, created, keyID)
	if err != nil {
		return ErrInvalidSignature
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return ErrInvalidSignature
	}

	valid := ed25519.Verify(opts.PublicKey, []byte(baseString), sigBytes)
	if !valid {
		return ErrInvalidSignature
	}

	return nil
}
