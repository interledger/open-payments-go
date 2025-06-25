package openpayments_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"

	"github.com/h2non/gock"
	openpayments "github.com/interledger/open-payments-go"
	was "github.com/interledger/open-payments-go/generated/walletaddressserver"
	"github.com/stretchr/testify/assert"
)

type walletAddressOption func(*was.WalletAddress)

func withAssetCode(assetCode string) walletAddressOption {
	return func(wa *was.WalletAddress) {
		wa.AssetCode = assetCode
	}
}

func withAssetScale(assetScale int) walletAddressOption {
	return func(wa *was.WalletAddress) {
		wa.AssetScale = assetScale
	}
}

func withAuthServer(authServer string) walletAddressOption {
	return func(wa *was.WalletAddress) {
		wa.AuthServer = &authServer
	}
}

func withID(id string) walletAddressOption {
	return func(wa *was.WalletAddress) {
		wa.Id = &id
	}
}

func withPublicName(publicName string) walletAddressOption {
	return func(wa *was.WalletAddress) {
		wa.PublicName = &publicName
	}
}

func withResourceServer(resourceServer string) walletAddressOption {
	return func(wa *was.WalletAddress) {
		wa.ResourceServer = &resourceServer
	}
}

func mockWalletAddress(options ...walletAddressOption) was.WalletAddress {
	defaultAuthServer := "https://auth.example.com"
	defaultID := "https://example.com/.well-known/pay"
	defaultResourceServer := "https://example.com"
	defaultPublicName := "Bob"

	walletAddress := was.WalletAddress{
		AssetCode:      "USD",
		AssetScale:     2,
		AuthServer:     &defaultAuthServer,
		Id:             &defaultID,
		PublicName:     &defaultPublicName,
		ResourceServer: &defaultResourceServer,
	}

	for _, option := range options {
		option(&walletAddress)
	}

	return walletAddress
}

var client *openpayments.Client
var httpClient *http.Client
var serverAddress = "https://example.com"

func init() {
	httpClient = &http.Client{
		Transport: http.DefaultTransport,
	}
	client = openpayments.NewClient(openpayments.WithHTTPClientUnauthed(httpClient))
}

func TestWalletAddressGet(t *testing.T) {
	defer gock.Off()
	gock.InterceptClient(httpClient)

	wa := mockWalletAddress()

	gock.New(serverAddress).
		Get("/.well-known/pay").
		Reply(200).
		JSON(wa)

	res, err := client.WalletAddress.Get(context.Background(), openpayments.WalletAddressGetParams{
		URL: fmt.Sprintf("%s%s", serverAddress, "/.well-known/pay"),
	})

	assert.NoError(t, err)
	assert.Equal(t, wa, res)
	assert.True(t, gock.IsDone())
}

type jsonWebKeyOption func(*was.JsonWebKey)

func withAlg(alg was.JsonWebKeyAlg) jsonWebKeyOption {
	return func(j *was.JsonWebKey) {
		j.Alg = alg
	}
}

func withCrv(crv was.JsonWebKeyCrv) jsonWebKeyOption {
	return func(j *was.JsonWebKey) {
		j.Crv = crv
	}
}

func withKid(kid string) jsonWebKeyOption {
	return func(j *was.JsonWebKey) {
		j.Kid = kid
	}
}

func withKty(kty was.JsonWebKeyKty) jsonWebKeyOption {
	return func(j *was.JsonWebKey) {
		j.Kty = kty
	}
}

func withUse(use was.JsonWebKeyUse) jsonWebKeyOption {
	return func(j *was.JsonWebKey) {
		j.Use = &use
	}
}

func withX(x string) jsonWebKeyOption {
	return func(j *was.JsonWebKey) {
		j.X = x
	}
}

func generateRandomID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func mockJSONWebKey(options ...jsonWebKeyOption) was.JsonWebKey {
	jwk := was.JsonWebKey{
		Alg: was.JsonWebKeyAlg("EdDSA"),
		Crv: was.JsonWebKeyCrv("Ed25519"),
		Kid: generateRandomID(),
		Kty: was.JsonWebKeyKty("OKP"),
		Use: nil,
		X:   generateRandomID(),
	}

	for _, option := range options {
		option(&jwk)
	}

	return jwk
}

func mockJSONWebKeySet(keys ...was.JsonWebKey) was.JsonWebKeySet {
	if len(keys) == 0 {
		defaultKey := mockJSONWebKey()
		keys = []was.JsonWebKey{defaultKey}
	}

	return was.JsonWebKeySet{
		Keys: &keys,
	}
}

func TestWalletAddressKeysGet(t *testing.T) {
	defer gock.Off()
	gock.InterceptClient(httpClient)

	jwks := mockJSONWebKeySet()

	fmt.Println(jwks)

	gock.New(serverAddress).
		Get("/.well-known/pay/jwks.json").
		Reply(200).
		JSON(jwks)

	res, err := client.WalletAddress.GetKeys(context.Background(), openpayments.WalletAddressGetKeysParams{
		URL: fmt.Sprintf("%s%s", serverAddress, "/.well-known/pay/jwks.json"),
	})

	assert.NoError(t, err)
	assert.Equal(t, jwks, res)
	assert.True(t, gock.IsDone())
}
