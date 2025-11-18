package openpayments

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/interledger/open-payments-go/httpsignatureutils"
)

type RequestDoer func(req *http.Request) (*http.Response, error)

type Client struct {
	httpClient      *http.Client
	WalletAddress   *WalletAddressService
	IncomingPayment *PublicIncomingPaymentService
}

// ClientOption is used to configure optional behavior for the Open Payments client.
type ClientOption func(*Client)

// WithHTTPClientUnauthed allows setting a custom HTTP client.
//
// WARNING: Use with care. Replacing the internal http.Client could break
// built-in behavior.
func WithHTTPClientUnauthed(c *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = c
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Transport: http.DefaultTransport,
		},
	}
	for _, opt := range opts {
		opt(c)
	}

	c.WalletAddress = &WalletAddressService{DoUnsigned: c.httpClient.Do}
	c.IncomingPayment = &PublicIncomingPaymentService{DoUnsigned: c.httpClient.Do}

	return c
}

type AuthenticatedClient struct {
	httpClient       *http.Client
	preSignHook      func(req *http.Request)
	postSignHook     func(req *http.Request)
	walletAddressUrl string /** The wallet address which the client will identify itself by */
	privateKey       ed25519.PrivateKey
	keyId            string
	WalletAddress    *WalletAddressService
	Grant            *GrantService
	IncomingPayment  *IncomingPaymentService
	Quote            *QuoteService
	Token            *TokenService
	OutgoingPayment  *OutgoingPaymentService
}

// AuthenticatedClientOption is used to configure optional behavior for the authenticated client.
type AuthenticatedClientOption func(*AuthenticatedClient)

// WithHTTPClientAuthed allows setting a custom HTTP client.
//
// WARNING: Use with care. Replacing the internal http.Client could break
// built-in behavior.
func WithHTTPClientAuthed(c *http.Client) AuthenticatedClientOption {
	return func(ac *AuthenticatedClient) {
		ac.httpClient = c
	}
}

func WithPreSignHook(hook func(req *http.Request)) AuthenticatedClientOption {
	return func(c *AuthenticatedClient) {
		c.preSignHook = hook
	}
}

func WithPostSignHook(hook func(req *http.Request)) AuthenticatedClientOption {
	return func(c *AuthenticatedClient) {
		c.postSignHook = hook
	}
}

// loadPrivateKey attempts to load a private key from either a file path or base64 string.
// It first checks if the input looks like a file path (contains '/' or '\' or has common extensions),
// and if that file exists. If so, it loads from file. Otherwise, it treats it as base64.
func loadPrivateKey(keyInput string) (ed25519.PrivateKey, error) {
	// Check if it looks like a file path and if the file exists
	if isLikelyFilePath(keyInput) {
		if _, err := os.Stat(keyInput); err == nil {
			// File exists, load from file
			return httpsignatureutils.LoadKey(keyInput)
		}
	}

	// Try as base64
	return httpsignatureutils.LoadBase64Key(keyInput)
}

// isLikelyFilePath returns true if the string looks like it could be a file path
func isLikelyFilePath(s string) bool {
	if strings.Contains(s, "/") || strings.Contains(s, "\\") {
		return true
	}
	
	if strings.HasSuffix(s, ".pem") || strings.HasSuffix(s, ".key") {
		return true
	}
	
	if strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") || strings.HasPrefix(s, ".\\") || strings.HasPrefix(s, "..\\") {
		return true
	}
	
	return false
}

// NewAuthenticatedClient creates a new authenticated client.
// The privateKeyOrPath parameter can be either:
//   - A file path to a PEM-encoded Ed25519 private key (e.g., "./keys/private.pem")
//   - A base64-encoded PEM string
//
// The function automatically detects which format is provided.
func NewAuthenticatedClient(walletAddressUrl string, privateKeyOrPath string, keyId string, opts ...AuthenticatedClientOption) (*AuthenticatedClient, error) {
	if len(walletAddressUrl) > 0 && walletAddressUrl[0] == '$' {
		return nil, fmt.Errorf("invalid wallet address: %q (cannot start with '$')", walletAddressUrl)
	}

	edKey, err := loadPrivateKey(privateKeyOrPath)
	if err != nil {
		return nil, fmt.Errorf("error loading private key: %w", err)
	}

	httpClient := &http.Client{
		Transport: http.DefaultTransport,
	}

	c := &AuthenticatedClient{
		httpClient:       httpClient,
		walletAddressUrl: walletAddressUrl,
		privateKey:       edKey,
		keyId:            keyId,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.WalletAddress = &WalletAddressService{DoUnsigned: c.httpClient.Do}
	c.IncomingPayment = &IncomingPaymentService{
		DoUnsigned: httpClient.Do,
		DoSigned:   c.DoSigned,
	}
	c.Grant = &GrantService{
		DoSigned: c.DoSigned,
		client:   c.walletAddressUrl,
	}
	c.Quote = &QuoteService{
		DoSigned: c.DoSigned,
	}
	c.Token = &TokenService{
		DoSigned: c.DoSigned,
	}
	c.OutgoingPayment = &OutgoingPaymentService{
		DoSigned: c.DoSigned,
	}

	return c, nil
}

func createContentDigest(body []byte) string {
	hash := sha512.Sum512(body)
	b64Hash := base64.StdEncoding.EncodeToString(hash[:])
	digest := fmt.Sprintf("sha-512=:%s:", b64Hash)
	return digest
}

func (c *AuthenticatedClient) DoSigned(req *http.Request) (*http.Response, error) {
	if c.preSignHook != nil {
		c.preSignHook(req)
	}

	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if err := req.Body.Close(); err != nil {
			return nil, fmt.Errorf("failed to close request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))

		contentDigest := createContentDigest(bodyBytes)

		req.Header.Set("Content-Digest", contentDigest)
	}

	sigHeaders, err := httpsignatureutils.CreateSignatureHeaders(httpsignatureutils.SignOptions{
		Request:    req,
		PrivateKey: c.privateKey,
		KeyID:      c.keyId,
	})
	if err != nil {
		return nil, err
	}

	req.Header.Set("Signature", fmt.Sprintf("sig1=:%s:", sigHeaders.Signature))
	req.Header.Set("Signature-Input", sigHeaders.SignatureInput)

	if c.postSignHook != nil {
		c.postSignHook(req)
	}

	return c.httpClient.Do(req)
}

// TODO: use or lose this DoSigned implementation. Did not like the (more or less) necessary side effect
// of CreateHeaders mutating the request. CreateHeaders has to add the content digest/length before signing.
// Could clone req in CreateHeaders but that seemed a bit heavy and potentially complicated. Alternatively
// could maybe just rename? SetSignatureHeaders? Just feels like maybe thats doing too much in httpsignatureutils.

// func (c *AuthenticatedClient) DoSigned(req *http.Request) (*http.Response, error) {
// 	headers, err := httpsignatureutils.CreateHeaders(httpsignatureutils.SignOptions{
// 		Request:    req,
// 		PrivateKey: c.privateKey,
// 		KeyID:      c.keyId,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// This aren't actually necessary to set because CreateHeaders does it.
// 	// req.Header.Set("Content-Length", headers.ContentLength)
// 	// req.Header.Set("Content-Digest", headers.ContentDigest)
// 	req.Header.Set("Signature", fmt.Sprintf("sig1=:%s:", headers.Signature))
// 	req.Header.Set("Signature-Input", headers.SignatureInput)

// 	return c.httpClient.Do(req)
// }
