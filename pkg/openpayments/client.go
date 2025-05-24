package openpayments

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"crypto/x509"
	"strings"
	"time"

	// "crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"io"
	"net/http"

	"github.com/dunglas/httpsfv"
	"github.com/interledger/open-payments-go-sdk/internal/lib"
	"github.com/interledger/open-payments-go-sdk/pkg/httpsignatureutils"
	"github.com/yaronf/httpsign"
)

type Client2 struct {
	httpClient *http.Client
	WalletAddress *WalletAddressRoutes
	IncomingPayment *IncomingPaymentRoutes
	// shouldnt actually work on unauthed client
	Grant *GrantRoutes
}

func NewClient2() *Client2 {
	httpClient := http.Client{Transport: &lib.HeaderTransport{Base: http.DefaultTransport}}
	return &Client2{
		httpClient: &httpClient,
		WalletAddress: &WalletAddressRoutes{httpClient: &httpClient},
		IncomingPayment: &IncomingPaymentRoutes{httpClient: &httpClient},
		Grant: &GrantRoutes{httpClient: &httpClient},
	}
}
// This implementation uses httpsfv.
// however, noticed it was producing slightly diff results than node client.
// "sha-512=\"T9vg7p...g==\"" vs "sha-512=:/LfvBez/1...lQ==:"

// func createContentDigest(body []byte) (string, error) {
// 	hash := sha512.Sum512(body)
//  /// i dont think this hashingis right... just use sum?
// 	b64Hash := base64.StdEncoding.EncodeToString(hash[:])

// 	dict := httpsfv.NewDictionary()
// 	dict.Add("sha-512", httpsfv.NewItem(b64Hash))

// 	text, err := httpsfv.Marshal(dict)
// 	if err != nil {
// 			return "", err
// 	}

// 	fmt.Printf("Content-Digest: %s\n", text)

// 	return text, nil
// }

type AuthenticatedClient2 struct {
	httpClient      *http.Client
	walletAddressUrl string  /** The wallet address which the client will identify itself by */
	privateKey      ed25519.PrivateKey
	keyId           string
	Grant           *GrantRoutes
}

func createContentDigest2(body []byte) (string, error) {
	hash := sha512.Sum512(body)
	b64Hash := base64.StdEncoding.EncodeToString(hash[:])

	digest := fmt.Sprintf("sha-512=:%s:", b64Hash)
	fmt.Printf("Content-Digest: %s\n", digest)
	return digest, nil
}

func NewAuthenticatedClient2(walletAddressUrl string, privateKey string, keyId string) *AuthenticatedClient2 {
	edKey, err := loadBase64Key(privateKey)
	if err != nil {
		fmt.Println("Error loading key:", err)
		return nil
	}

	httpClient := http.Client{
		Transport: &lib.HeaderTransport{
			Base: http.DefaultTransport,
			// TODO: wrap more of this logic in a createHeaders or similar in httpsignatureutils? like `createHeaders` in the
			// TS version of http-signature-utils. then this woul dlooke like signRequest in the TS open payments client
			// - in requests.ts: https://github.com/interledger/open-payments/blob/main/packages/open-payments/src/client/requests.ts
			SigningFunc: func(req *http.Request) error {
				// Set headers to match the working request
				// req.Header.Set("Accept", "application/json, text/plain, */*")
				// req.Header.Set("Accept-Encoding", "gzip, compress, deflate, br")

				// TODO: REMOVE. stubbing this in for debugging purposes
				// jsonBody := "{\"access_token\":{\"access\":[{\"type\":\"incoming-payment\",\"actions\":[\"create\",\"read\",\"list\",\"complete\"]}]},\"client\":\"https://happy-life-bank-backend/accounts/pfry\"}"
				// req.Body = io.NopCloser(bytes.NewReader([]byte(jsonBody)))

				fmt.Println("req.ContentLength:", req.ContentLength)

				// Handle body and content digest
				if req.Body != nil {
					// Read and replace the body
					bodyBytes, err := io.ReadAll(req.Body)
					if err != nil {
							return err
					}
					req.Body.Close()
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					// TODO: may need to set length from body - although in this context we'll have it already on req
					// req.ContentLength = int64(len(bodyBytes))
					req.Header.Set("Content-Length", fmt.Sprintf("%d", req.ContentLength))

					contentDigest, err := createContentDigest2(bodyBytes)
					if err != nil {
						return err
					}
					req.Header.Set("Content-Digest", contentDigest)
				}

				sigHeaders, err := httpsignatureutils.CreateSignatureHeaders(httpsignatureutils.SignOptions{
					Request:    req,
					PrivateKey: edKey,
					KeyID:      keyId,
				})

				if err != nil {
					return err
				}

				// req.Header.Set("Signature", sigHeaders.Signature)
				// Matches what i see in validateSignature from good bruno request
				req.Header.Set("Signature", fmt.Sprintf("sig1=:%s:", sigHeaders.Signature))
				req.Header.Set("Signature-Input", sigHeaders.SignatureInput)

				// print the request here for debugging
				// fmt.Println("Request Headers:", req.Header)
				// fmt.Println("Request Body:", req.GetBody())
				// fmt.Println("Request URL:", req.URL)
				// fmt.Println("Request Method:", req.Method)
				// fmt.Println("Request Content-Length:", req.ContentLength)

				return nil
			},
		},
	}

	return &AuthenticatedClient2{
		httpClient: &httpClient,
		walletAddressUrl: walletAddressUrl,
		privateKey: edKey, //privateKey.()
		keyId: keyId,
		Grant: &GrantRoutes{httpClient: &httpClient},
	}
}

// Manual implemetnation of request signing
func NewAuthenticatedClient_man1(walletAddressUrl string, privateKey string, keyId string) *AuthenticatedClient2 {
	edKey, err := loadBase64Key(privateKey)
	if err != nil {
		fmt.Println("Error loading key:", err)
		return nil
	}

	httpClient := http.Client{
		Transport: &lib.HeaderTransport{
			Base: http.DefaultTransport,
			SigningFunc: func(req *http.Request) error {
				// Set headers to match the working request
				req.Header.Set("Accept", "application/json, text/plain, */*")
				req.Header.Set("Accept-Encoding", "gzip, compress, deflate, br")

				// Handle body and content digest
				if req.Body != nil {
					// bodyBytes, err := io.ReadAll(req.Body)
					// if err != nil {
					// 	return err
					// }

					// TODO: rm this. hardcoding for debugging purposes... ensure its identical to burno req
					bodyBytes := []byte("{\"access_token\":{\"access\":[{\"type\":\"incoming-payment\",\"actions\":[\"create\",\"read\",\"list\",\"complete\"]}]},\"client\":\"https://happy-life-bank-backend/accounts/pfry\"}")
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					// Set the correct content length
					req.ContentLength = int64(len(bodyBytes))
					req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))

					contentDigest, err := createContentDigest2(bodyBytes)
					if err != nil {
						return err
					}
					req.Header.Set("Content-Digest", contentDigest)
				}

				components := createSignatureComponents(req)

				// Create signature input string with alg parameter
				// created := time.Now().Unix()
				// unhardcode this - why is the signature different from bruno? created, signature input, digest are same
				created := 1747856886
				// created := 1747797344
				// does 1747797344 match this signature?
				// happy-life-auth-1      |       signature: 'sig1=:non+QR39gdeGirG2Dsyn0AKyREA2llfdnY0v7fdZlT0rCNtp6ARYYfntPTbPxOarQ2lE7apI7bEmqK9I+IHDBg==:',
				// happy-life-auth-1      |       'signature-input': 'sig1=("@method" "@target-uri" "content-digest" "content-length" "content-type");created=1747797344;keyid="keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5";alg="ed25519"',
				// no... why
				signatureInput := fmt.Sprintf(
					// `sig1=(%s);keyid="%s";alg="ed25519";created=%d`,
					`sig1=("%s");created=%d;keyid="%s";alg="ed25519"`,
					strings.Join(components, `" "`),
					created,
					keyId,
				)

				host := req.Host
				if host == "" {
					host = req.URL.Host
				}
				targetUri := req.URL.Scheme + "://" + host + req.URL.RequestURI()

				// Create string to sign
				toSign := fmt.Sprintf(
					"@method: %s\n@target-uri: %s",
					req.Method,
					// req.URL.String(),
					targetUri,
				)

				// Add optional components
				if req.Header.Get("Authorization") != "" {
					toSign += fmt.Sprintf("\nauthorization: %s", req.Header.Get("Authorization"))
				}
				if req.Body != nil {
					toSign += fmt.Sprintf("\ncontent-digest: %s", req.Header.Get("Content-Digest"))
					toSign += fmt.Sprintf("\ncontent-length: %s", req.Header.Get("Content-Length"))
					toSign += fmt.Sprintf("\ncontent-type: %s", req.Header.Get("Content-Type"))
				}

				fmt.Println("toSign", toSign)

				// Sign the message
				signature := ed25519.Sign(edKey, []byte(toSign))
				signatureString := fmt.Sprintf("sig1=:%s:", base64.StdEncoding.EncodeToString(signature))

				fmt.Println("signedSignature", signatureString)
				// this is always the same for the same request... whereas bruno/lib implementation are different (timestamp?)

				req.Header.Set("Signature", signatureString)
				// req.Header.Set("Signature", "icbZKtQhY2BbjpNXQaCBfYJRNJ675vWUZA4UUfDUoGahC5nCYWvnh6b+5h6dE17nnX+yw7QVijM06gfyBmhlC")
				req.Header.Set("Signature-Input", signatureInput)

				return nil
			},
		},
	}

	return &AuthenticatedClient2{
		httpClient: &httpClient,
		walletAddressUrl: walletAddressUrl,
		privateKey: edKey, //privateKey.()
		keyId: keyId,
		Grant: &GrantRoutes{httpClient: &httpClient},
	}
}

// Implemenation using httpsign for request signing
// func NewAuthenticatedClient(walletAddressUrl, privateKey, keyId string) *AuthenticatedClient {
func NewAuthenticatedClient_httpsign(walletAddressUrl, privateKey, keyId string) *AuthenticatedClient2 {
	// for generating a key in go and passing in
	// pemBytes, err := base64.StdEncoding.DecodeString(privateKey)
	// if err != nil {
	// 	panic(fmt.Sprintf("Invalid base64 private key: %v", err))
	// }
	// edKey := ed25519.PrivateKey(pemBytes) // Simplified key extraction

	// for using what we have in bruno config
	edKey, err := loadBase64Key(privateKey)
	if err != nil {
		fmt.Println("Error loading key:", err)
		return nil
	}

	// create signer
	signConfig := httpsign.NewSignConfig().SetKeyID(keyId)
	fields := httpsign.NewFields().AddHeaders("@method", "@target-uri", "content-digest", "content-length", "content-type")
	signer, err := httpsign.NewEd25519Signer(edKey, signConfig, *fields)
	if err != nil {
		panic(fmt.Sprintf("Failed to create signer: %v", err))
	}

	// custom RoundTripper for signing requests
	signingTransport := &SigningRoundTripper{
		Transport: http.DefaultTransport,
		Signer:    signer,
	}

	httpClient := &http.Client{
		Transport: signingTransport,
	}

	return &AuthenticatedClient2{
		httpClient:      httpClient,
		walletAddressUrl: walletAddressUrl,
		privateKey:      edKey,
		keyId:           keyId,
		Grant:           &GrantRoutes{httpClient: httpClient},
	}
}

// Custom RoundTripper to sign requests
type SigningRoundTripper struct {
	Transport http.RoundTripper
	Signer    *httpsign.Signer
}

// NOT WORKING
// used by httpsign
// I have some suspicion that the handling of the body is off but attempts to make it work
// like working version doesnt yield any success
func (s *SigningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil && req.ContentLength > 0 {

		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		// hardcode in body that we see bruno producing for testing purposes
		// bodyBytes := []byte("{\"access_token\":{\"access\":[{\"type\":\"incoming-payment\",\"actions\":[\"create\",\"read\",\"list\",\"complete\"]}]},\"client\":\"https://happy-life-bank-backend/accounts/pfry\"}")
		// ^ produces the same hash as bruno. still get invalid signature
		// might be onto something though... content digest value matches bruno for the first time.
		// - 'sha-512=:/LfvBez/1knzYV3v4+Ej1qidX28IuoPp4jJBNSTkgBAu5TN5qS2FrfEWJohbBjIk1Xg7+qanR6VPm2+XyrZ3lQ==:'

		// get content length
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		req.ContentLength = int64(len(bodyBytes))
		req.Header.Set("Content-length", string(len(bodyBytes)))
		bodyReader := io.NopCloser(bytes.NewReader(bodyBytes))

		contentDigest, err := httpsign.GenerateContentDigestHeader(&bodyReader, []string{"sha-512"})
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Digest", contentDigest)
	}

	signatureInput, signature, err := httpsign.SignRequest("sig1", *s.Signer, req)
	if err != nil {
		return nil, err
	}

	// the signature is different each time. so is bruno. my manual one isnt. why?
	fmt.Println("signedSignature:", signature)

	req.Header.Set("Signature", signature)
	req.Header.Set("Signature-Input", signatureInput)

	// Send the request
	return s.Transport.RoundTrip(req)
}

func loadBase64Key(base64Key string) (ed25519.PrivateKey, error) {
	// base64 key in bruno decoded should be:
	// -----BEGIN PRIVATE KEY-----
	// MC4CAQAwBQYDK2VwBCIEIEqezmcPhOE8bkwN+jQrppfRYzGIdFTVWQGTHJIKpz88
	// -----END PRIVATE KEY-----

	pemBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	// Step : Decode the PEM block
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	blockJSON, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JWK:", err)
		return nil, fmt.Errorf("could not marshal json")
	}
	fmt.Println("Parsed key from PEM file:", string(blockJSON))

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	if edKey, ok := privateKey.(ed25519.PrivateKey); ok {
		fmt.Printf("Ed25519 Private Key: %x\n", edKey)
		return edKey, nil
	} else {
		return nil, fmt.Errorf("key is not an Ed25519 key")
	}
}

// func loadBase64Key(keyBase64 string) (ed25519.PrivateKey, error) {
// 	pemBytes, err := base64.StdEncoding.DecodeString(keyBase64)
// 	if err != nil {
// 		return nil, fmt.Errorf("Invalid base64: %v", err)
// 	}

// 	key, err := x509.ParsePKCS8PrivateKey(pemBytes)
// 	if err != nil {
// 		return nil, fmt.Errorf("ParsePKCS8PrivateKey failed: %v", err)
// 	}

// 	edKey, ok := key.(ed25519.PrivateKey)
// 	if !ok {
// 		return nil, fmt.Errorf("Not an Ed25519 private key")
// 	}
// 	return edKey, nil
// }

func createSignatureComponents(req *http.Request) []string {
	components := []string{"@method", "@target-uri"}

	if req.Header.Get("Authorization") != "" {
		components = append(components, "authorization")
	}

	if req.Body != nil {
		components = append(components, "content-digest", "content-length", "content-type")
	}

	return components
}

// SigningRoundTripper is a custom http.RoundTripper that signs outgoing requests
type HTTPSFVSigningRoundTripper struct {
	Transport  http.RoundTripper
	PrivateKey ed25519.PrivateKey
	KeyID      string
}

func (s *HTTPSFVSigningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Handle body and Content-Digest if present
	if req.Body != nil && req.ContentLength > 0 {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}

		// Restore the body for the actual request
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		req.ContentLength = int64(len(bodyBytes))
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))

		// Generate Content-Digest using httpsfv
		contentDigest, err := createContentDigestWithHttpsfv(bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to create content digest: %w", err)
		}
		req.Header.Set("Content-Digest", contentDigest)
	}

	// Create and set signature headers
	signatureInput, signature, err := signRequestWithHttpsfv(req, s.PrivateKey, s.KeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	req.Header.Set("Signature-Input", signatureInput)
	req.Header.Set("Signature", signature)

	// Forward the signed request to the underlying transport
	return s.Transport.RoundTrip(req)
}

// Create Content-Digest header using httpsfv
func createContentDigestWithHttpsfv(body []byte) (string, error) {
	// Calculate SHA-512 hash of the body
	hash := sha512.Sum512(body)
	b64Hash := base64.StdEncoding.EncodeToString(hash[:])

	// dict := httpsfv.NewDictionary()
	// dict.Add("sha-512", httpsfv.NewItem(b64Hash))

	// Create a dictionary with a single key "sha-512" and the base64 hash as its value
	dict := httpsfv.NewDictionary()
	dict.Add("sha-512", httpsfv.NewItem(b64Hash))

	// Serialize the dictionary to a string
	serialized, err := httpsfv.Marshal(dict)
	if err != nil {
		return "", fmt.Errorf("failed to marshal content digest: %w", err)
	}

	return serialized, nil
}

// Sign request using httpsfv
func signRequestWithHttpsfv(req *http.Request, privateKey ed25519.PrivateKey, keyID string) (string, string, error) {
	// Determine which components to include in the signature
	components := []string{"@method", "@target-uri"}

	// Add optional components if they exist in the headers
	if req.Header.Get("Authorization") != "" {
		components = append(components, "authorization")
	}

	if req.Body != nil {
		components = append(components, "content-digest", "content-length", "content-type")
	}

	// Create signature input parameters
	created := time.Now().Unix()

	// Build signature input dictionary using httpsfv
	signatureInputDict := httpsfv.NewDictionary()

	// Create inner list with items for each component
	innerList := httpsfv.InnerList{
		Items:  make([]httpsfv.Item, len(components)),
		Params: httpsfv.NewParams(),
	}

	// Add component identifiers to the inner list
	for i, component := range components {
		innerList.Items[i] = httpsfv.NewItem(component)
	}

	// Add signature parameters
	innerList.Params.Add("created", created)
	innerList.Params.Add("keyid", keyID)
	innerList.Params.Add("alg", "ed25519")

	// Add the inner list to the dictionary with key "sig1"
	signatureInputDict.Add("sig1", &innerList)

	// Serialize the signature input
	signatureInputStr, err := httpsfv.Marshal(signatureInputDict)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal signature input: %w", err)
	}

	// Construct the string to sign
	toSign := buildStringToSign(req, components)

	// Sign the message
	signature := ed25519.Sign(privateKey, []byte(toSign))

	// Create signature dictionary
	signatureDict := httpsfv.NewDictionary()
	signatureDict.Add("sig1", httpsfv.NewItem(base64.StdEncoding.EncodeToString(signature)))

	// Serialize the signature
	signatureStr, err := httpsfv.Marshal(signatureDict)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal signature: %w", err)
	}

	return signatureInputStr, signatureStr, nil
}

// Build the string to sign according to the HTTP Message Signatures spec
func buildStringToSign(req *http.Request, components []string) string {
	var builder strings.Builder

	for i, component := range components {
		if i > 0 {
			builder.WriteString("\n")
		}

		switch component {
		case "@method":
			builder.WriteString("@method: ")
			builder.WriteString(req.Method)
		case "@target-uri":
			builder.WriteString("@target-uri: ")
			builder.WriteString(req.URL.String())
		default:
			// For regular headers, use lowercase name
			headerName := strings.ToLower(component)
			headerValue := req.Header.Get(headerName)
			if headerValue != "" {
				builder.WriteString(headerName)
				builder.WriteString(": ")
				builder.WriteString(headerValue)
			}
		}
	}

	return builder.String()
}

// NewAuthenticatedClientWithHttpsfv creates a new client that uses httpsfv for structured headers
// func NewAuthenticatedClient(walletAddressUrl, privateKey, keyId string) *AuthenticatedClient {
func NewAuthenticatedClientWithHttpsfv(walletAddressUrl, privateKey, keyId string) *AuthenticatedClient2 {

	edKey, err := loadBase64Key(privateKey)
	if err != nil {
		fmt.Println("Error loading key:", err)
		return nil
	}

	httpClient := &http.Client{
		Transport: &HTTPSFVSigningRoundTripper{
			Transport:  http.DefaultTransport,
			PrivateKey: edKey,
			KeyID:      keyId,
		},
	}

	return &AuthenticatedClient2{
		httpClient:       httpClient,
		walletAddressUrl: walletAddressUrl,
		privateKey:       edKey,
		keyId:            keyId,
		Grant:            &GrantRoutes{httpClient: httpClient},
	}
}