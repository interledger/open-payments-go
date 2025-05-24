package httpsignatureutils

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrMissingSignature     = errors.New("missing signature")
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
			components = strings.Fields(inner)
			
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



func ValidateSignature(opts *ValidationOptions) error {
	sigInput := opts.Headers.Get("Signature-Input")
	if sigInput == "" {
		return ErrMissingSignatureInput
	}

	components, created, keyID, err := parseSignatureInput(sigInput)
	if err != nil {
		return err
	}

	sig := opts.Headers.Get("Signature")
	if sig == "" {
		return ErrMissingSignature
	}

	baseString := createSignatureBaseString(opts.Request, components, created, keyID)

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
