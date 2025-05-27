package httpsignatureutils

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type SignatureHeaders struct {
    Signature      string
    SignatureInput string
}

// TODO: use error if createSignatureBaseString cant find header for a given component?
// var SignatureError = errors.New("failed to create signature")

type SignOptions struct {
    Request    *http.Request
    PrivateKey ed25519.PrivateKey
    KeyID      string
}

func createSignatureBaseString(req *http.Request, components []string, created int64, keyID string) string {
    var parts []string

    for _, comp := range components {
        var value string
        switch strings.ToLower(comp) {
        case "@method":
            value = req.Method
        case "@target-uri":
            urlStr := req.URL.String()
            if !strings.HasSuffix(urlStr, "/") {
                urlStr += "/"
            }
            value = urlStr
        case "authorization":
            if auth := req.Header.Get("Authorization"); auth != "" {
                value = auth
            }
        case "content-digest":
            if cd := req.Header.Get("Content-Digest"); cd != "" {
                value = cd
            }
        case "content-length":
            if cl := req.Header.Get("Content-Length"); cl != "" {
                value = cl
            }
        case "content-type":
            if ct := req.Header.Get("Content-Type"); ct != "" {
                value = ct
            }
        default:
            value = ""
        }
        parts = append(parts, fmt.Sprintf("\"%s\": %s", comp, value))
    }

    quotedComponents := make([]string, len(components))
    for i, comp := range components {
        quotedComponents[i] = fmt.Sprintf("\"%s\"", comp)
    }
    sigParams := fmt.Sprintf("(%s);created=%d;keyid=\"%s\";alg=\"ed25519\"", strings.Join(quotedComponents, " "), created, keyID)

    parts = append(parts, fmt.Sprintf("\"@signature-params\": %s", sigParams))

    return strings.Join(parts, "\n")
}

func CreateSignatureHeaders(opts SignOptions) (*SignatureHeaders, error) {
    components := []string{"@method", "@target-uri", "content-type"}

    if opts.Request.Header.Get("Authorization") != "" {
        components = append(components, "authorization")
    }

    if opts.Request.ContentLength > 0 {
        components = append(components, "content-digest", "content-length")
    }

    created := time.Now().Unix()

    signatureBase := createSignatureBaseString(opts.Request, components, created, opts.KeyID)

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
