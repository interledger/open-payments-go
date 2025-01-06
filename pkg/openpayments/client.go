package openpayments

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
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

	"github.com/interledger/open-payments-go-sdk/internal/lib"
	"github.com/yaronf/httpsign"
)

type Client struct {
	httpClient *http.Client
	WalletAddress *WalletAddressRoutes
	IncomingPayment *IncomingPaymentRoutes
	// shouldnt actually work on unauthed client
	Grant *GrantRoutes
}

func NewClient() *Client {
	httpClient := http.Client{Transport: &lib.HeaderTransport{Base: http.DefaultTransport}}
	return &Client{
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

type AuthenticatedClient struct {
	httpClient      *http.Client
	walletAddressUrl string
	privateKey      ed25519.PrivateKey
	keyId           string
	Grant           *GrantRoutes
}

// Manual implemetnation of request signing
func NewAuthenticatedClient_man(walletAddressUrl string, privateKey string, keyId string) *AuthenticatedClient {
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
					
					// hardcoding for testing purposes... ensure its identical to burno req
					bodyBytes := []byte("{\"access_token\":{\"access\":[{\"type\":\"incoming-payment\",\"actions\":[\"create\",\"read\",\"list\",\"complete\"]}]},\"client\":\"https://happy-life-bank-backend/accounts/pfry\"}")
					req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					// Set the correct content length
					req.ContentLength = int64(len(bodyBytes))
					req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))

					contentDigest, err := createContentDigest(bodyBytes)
					if err != nil {
						return err
					}
					req.Header.Set("Content-Digest", contentDigest)
				}

				components := createSignatureComponents(req)

				// Create signature input string with alg parameter
				created := time.Now().Unix()
				signatureInput := fmt.Sprintf(
					// `sig1=(%s);keyid="%s";alg="ed25519";created=%d`,
					`sig1=("%s");created=%d;keyid="%s";alg="ed25519"`,
					strings.Join(components, `" "`),
					created,
					keyId,
				)

				// Create string to sign
				toSign := fmt.Sprintf(
					"@method: %s\n@target-uri: %s",
					req.Method,
					req.URL.String(),
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
				req.Header.Set("Signature-Input", signatureInput)

				return nil
			},
		},
	}

	return &AuthenticatedClient{
		httpClient: &httpClient,
		walletAddressUrl: walletAddressUrl,
		privateKey: edKey, //privateKey.()
		keyId: keyId,
		Grant: &GrantRoutes{httpClient: &httpClient},
	}
}

func createContentDigest(body []byte) (string, error) {
	hash := sha512.Sum512(body)
	b64Hash := base64.StdEncoding.EncodeToString(hash[:])

	digest := fmt.Sprintf("sha-512=:%s:", b64Hash)
	fmt.Printf("Content-Digest: %s\n", digest)
	return digest, nil
}

// Implemenation using httpsign for request signing
func NewAuthenticatedClient(walletAddressUrl, privateKey, keyId string) *AuthenticatedClient {
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

	return &AuthenticatedClient{
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

func (s *SigningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil && req.ContentLength > 0 {

		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		// hardcode in body that we see bruno producing for testing purposes
		// bodyBytes := []byte("{\"access_token\":{\"access\":[{\"type\":\"incoming-payment\",\"actions\":[\"create\",\"read\",\"list\",\"complete\"]}]},\"client\":\"https://happy-life-bank-backend/accounts/pfry\"}")
		// ^ produces the same hash as bruno. still get invalid signature

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

	///////////////////////
	// TODO: rm... just testing key details in rafiki. can it verify?
	// - no, it fails... suggests something is wrong with the private key im signing with here
	jwk := JWK{
		Kid: "keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5",
		X:   "ubqoInifJ5sssIPPnQR1gVPfmoZnJtPhTkyMXNoJF_8",
		Alg: "EdDSA",
		Kty: "OKP",
		Crv: "Ed25519",
	}
	// Perform verification
	valid, err := verifyRequest(req, jwk)
	if err != nil {
		fmt.Println("Verification failed:", err)
	} else if valid {
		fmt.Println("Signature is valid!")
	} else {
		// this is firing... means i have a problem with the key Im signign with i guess
		fmt.Println("Signature is invalid.")
	}
	////////////////////////


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

// TESTING that the signed request can be verified using the client key details in rafiki

type JWK struct {
	Kid string `json:"kid"`
	X   string `json:"x"`
	Alg string `json:"alg"`
	Kty string `json:"kty"`
	Crv string `json:"crv"`
}

// Function to verify the request
func verifyRequest(req *http.Request, jwk JWK) (bool, error) {
	// Decode the base64 "x" value to get the public key
	publicKeyBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return false, fmt.Errorf("error decoding public key: %v", err)
	}

	// Read the signature from the headers
	signature := req.Header.Get("Signature")
	if signature == "" {
		return false, fmt.Errorf("signature missing in headers")
	}

	// Split on colons and get the base64-encoded signature part
	parts := strings.Split(signature, ":")
	if len(parts) != 3 {
		fmt.Println("Invalid signature format")
		return false, fmt.Errorf("could not parse signature from header")
	}

	signatureB64 := parts[1] // This is the base64-encoded signature
	fmt.Println("Extracted Base64 Signature:", signatureB64)

	// Reconstruct the signed data
	signedData, err := reconstructSignedData(req)
	if err != nil {
		return false, fmt.Errorf("failed to reconstruct signed data: %v", err)
	}

	// Decode the signature from base64
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		fmt.Println(signature)
		return false, fmt.Errorf("error decoding signature: %v", err)
	}

	// Verify the signature using Ed25519
	isValid := ed25519.Verify(ed25519.PublicKey(publicKeyBytes), signedData, signatureBytes)
	return isValid, nil
}

// Function to reconstruct signed data from request (simplified for demo)
func reconstructSignedData(req *http.Request) ([]byte, error) {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore the body
	}
	// Include method, target URI, and body in signature input (example logic)
	return []byte(fmt.Sprintf("%s %s %s", req.Method, req.URL.String(), bodyBytes)), nil
}

// TODO: rm if unused
func generateJSONWebKey(){
		// Generate Ed25519 key pair
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			fmt.Println("Error generating key:", err)
			return
		}
	
		// Convert the public key to base64 URL encoding
		x := base64.RawURLEncoding.EncodeToString(publicKey)
	
		// Create JWK object
		jwk := JWK{
			Kid: "keyid-123", // Replace with your key ID
			X:   x,
			Alg: "EdDSA",
			Kty: "OKP",
			Crv: "Ed25519",
		}
	
		// Convert JWK object to JSON
		jwkJSON, err := json.MarshalIndent(jwk, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling JWK:", err)
			return
		}
	
		// Print the JWK
		fmt.Println(string(jwkJSON))
	
		// Optional: Export the private key for later use
		privateKeyBase64 := base64.RawURLEncoding.EncodeToString(privateKey.Seed())
		fmt.Println("Private Key (base64):", privateKeyBase64)
}
// TODO: rm if unused
func GenerateNewPrivateKey() (string, error) {
	// Generate a new Ed25519 key pair
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", err
	}

	encodedKey := base64.StdEncoding.EncodeToString(privateKey)

	return encodedKey, nil
}