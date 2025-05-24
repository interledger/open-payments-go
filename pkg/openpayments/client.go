package openpayments

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/interledger/open-payments-go-sdk/internal/lib"
	"github.com/interledger/open-payments-go-sdk/pkg/httpsignatureutils"
)

type RequestDoer func(req *http.Request) (*http.Response, error)

type Client struct {
	httpClient *http.Client
	WalletAddress *WalletAddressRoutes
	IncomingPayment *PublicIncomingPaymentService
}

func NewClient() *Client {
	httpClient := http.Client{Transport: &lib.HeaderTransport{Base: http.DefaultTransport}}
	return &Client{
		httpClient: &httpClient,
		WalletAddress: &WalletAddressRoutes{httpClient: &httpClient},
		IncomingPayment: &PublicIncomingPaymentService{DoUnsigned: httpClient.Do},
	}
}

type AuthenticatedClient struct {
	httpClient      *http.Client
	walletAddressUrl string  /** The wallet address which the client will identify itself by */
	privateKey      ed25519.PrivateKey
	keyId           string
	Grant           *GrantService
	IncomingPayment *IncomingPaymentService
}

func createContentDigest(body []byte) (string) {
	hash := sha512.Sum512(body)
	b64Hash := base64.StdEncoding.EncodeToString(hash[:])
	digest := fmt.Sprintf("sha-512=:%s:", b64Hash)
	return digest
}

func NewAuthenticatedClient(walletAddressUrl string, privateKey string, keyId string) *AuthenticatedClient {
	edKey, err := httpsignatureutils.LoadBase64Key(privateKey)
	if err != nil {
		fmt.Println("Error loading key:", err)
		return nil
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
	c.IncomingPayment = &IncomingPaymentService{
		DoUnsigned: httpClient.Do,
		DoSigned:   c.DoSigned,
	}
	c.Grant = &GrantService{
		DoSigned:   c.DoSigned,
	}

	return c
}

func (c *AuthenticatedClient) DoSigned(req *http.Request) (*http.Response, error) {
	// Read and re-insert body if present
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
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

	return c.httpClient.Do(req)
}
