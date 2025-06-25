package httpsignatureutils

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrMissingRequiredHeader = errors.New("missing required header")
)

type ContentAndSignatureHeaders struct {
    Signature      string
    SignatureInput string
    ContentDigest  string
    ContentLength  string
}

type SignatureHeaders struct {
    Signature      string
    SignatureInput string
}

type SignOptions struct {
    Request    *http.Request
    PrivateKey ed25519.PrivateKey
    KeyID      string
}

func createSignatureBaseString(req *http.Request, components []string, created int64, keyID string) (string, error) {
    var parts []string

    for _, comp := range components {
        var value string
        switch strings.ToLower(comp) {
        case "@method":
            value = req.Method
        case "@target-uri":
            urlStr := req.URL.String()
            // Add a trailing slash only if there's no path component
            if req.URL.Path == "" {
                urlStr += "/"
            }
            value = urlStr
        case "authorization":
            value = req.Header.Get("Authorization")
        case "content-digest":
            value = req.Header.Get("Content-Digest")
        case "content-length":
            value = req.Header.Get("Content-Length")
        case "content-type":
            value = req.Header.Get("Content-Type")
        default:
            // try to get any other component as a header.
            value = req.Header.Get(comp)
            if value == "" {
                return "", fmt.Errorf("%w: %s", ErrMissingRequiredHeader, comp)
            }
        }
        parts = append(parts, fmt.Sprintf("\"%s\": %s", comp, value))
    }

    quotedComponents := make([]string, len(components))
    for i, comp := range components {
        quotedComponents[i] = fmt.Sprintf("\"%s\"", comp)
    }
    sigParams := fmt.Sprintf("(%s);created=%d;keyid=\"%s\";alg=\"ed25519\"", strings.Join(quotedComponents, " "), created, keyID)

    parts = append(parts, fmt.Sprintf("\"@signature-params\": %s", sigParams))

    return strings.Join(parts, "\n"), nil
}

func CreateSignatureHeaders(opts SignOptions) (*SignatureHeaders, error) {
    components := []string{"@method", "@target-uri"}

    if opts.Request.Header.Get("Authorization") != "" || opts.Request.Header.Get("authorization") != "" {
        components = append(components, "authorization")
    }

    if opts.Request.ContentLength > 0 {
        components = append(components, "content-digest", "content-length", "content-type")
    }

    created := time.Now().Unix()

    signatureBase, err := createSignatureBaseString(opts.Request, components, created, opts.KeyID)

	if err != nil {
		return nil, fmt.Errorf("failed to create signature base string: %w", err)
	}

    signatureBytes := ed25519.Sign(opts.PrivateKey, []byte(signatureBase))
    signature := base64.StdEncoding.EncodeToString(signatureBytes)

    quotedComponents := make([]string, len(components))
    for i, comp := range components {
        quotedComponents[i] = fmt.Sprintf("\"%s\"", comp)
    }
    signatureInput := fmt.Sprintf("sig1=(%s);created=%d;keyid=\"%s\";alg=\"ed25519\"", strings.Join(quotedComponents, " "), created, opts.KeyID)

    return &SignatureHeaders{
        Signature:      signature,
        SignatureInput: signatureInput,
    }, nil
}


func createContentDigest(body []byte) (string) {
	hash := sha512.Sum512(body)
	b64Hash := base64.StdEncoding.EncodeToString(hash[:])
	digest := fmt.Sprintf("sha-512=:%s:", b64Hash)
	return digest
}

func CreateHeaders(opts SignOptions) (*ContentAndSignatureHeaders, error) {
    req := opts.Request
    var bodyBytes []byte
    if req.Body != nil {
        var err error
        bodyBytes, err = io.ReadAll(req.Body)
        if err != nil {
            return nil, err
        }
        req.Body.Close()
        req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
    }

    if len(bodyBytes) > 0 {
		req.Header.Set("Content-Digest", createContentDigest(bodyBytes))
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
    }

    sigHeaders, err := CreateSignatureHeaders(opts)
    if err != nil {
        return nil, err
    }

    return &ContentAndSignatureHeaders{
		Signature:      sigHeaders.Signature,
		SignatureInput: sigHeaders.SignatureInput,
		ContentDigest:  req.Header.Get("Content-Digest"),
		ContentLength:  req.Header.Get("Content-Length"),
    }, nil
}
