package testutils

import (
	"crypto/rand"
	"encoding/hex"

	was "github.com/interledger/open-payments-go/generated/walletaddressserver"
)

type MockWalletAddressBuilder struct {
	wa was.WalletAddress
}

func NewMockWalletAddressBuilder() *MockWalletAddressBuilder {
	defaultAuthServer := "https://auth.example.com"
	defaultID := "https://example.com/.well-known/pay"
	defaultResourceServer := "https://example.com"
	defaultPublicName := "Bob"

	return &MockWalletAddressBuilder{
		wa: was.WalletAddress{
			Id:             &defaultID,
			PublicName:     &defaultPublicName,
			AssetScale:     2,
			AssetCode:      "USD",
			AuthServer:     &defaultAuthServer,
			ResourceServer: &defaultResourceServer,
		},
	}
}

func (b *MockWalletAddressBuilder) WithAssetCode(assetCode string) *MockWalletAddressBuilder {
	b.wa.AssetCode = assetCode
	return b
}

func (b *MockWalletAddressBuilder) WithAssetScale(assetScale int) *MockWalletAddressBuilder {
	b.wa.AssetScale = assetScale
	return b
}

func (b *MockWalletAddressBuilder) WithAuthServer(authServer string) *MockWalletAddressBuilder {
	b.wa.AuthServer = &authServer
	return b
}

func (b *MockWalletAddressBuilder) WithID(id string) *MockWalletAddressBuilder {
	b.wa.Id = &id
	return b
}

func (b *MockWalletAddressBuilder) WithPublicName(publicName string) *MockWalletAddressBuilder {
	b.wa.PublicName = &publicName
	return b
}

func (b *MockWalletAddressBuilder) WithResourceServer(resourceServer string) *MockWalletAddressBuilder {
	b.wa.ResourceServer = &resourceServer
	return b
}

func (b *MockWalletAddressBuilder) Build() was.WalletAddress {
	return b.wa
}

type MockJWKBuilder struct {
	jwk was.JsonWebKey
}

func generateRandomID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func NewMockJWKBuilder() *MockJWKBuilder {
	defaultKid := generateRandomID()
	defaultX := generateRandomID()

	return &MockJWKBuilder{
		jwk: was.JsonWebKey{
			Kid: defaultKid,
			X:   defaultX,
			Use: nil,
			Alg: was.JsonWebKeyAlg("EdDSA"),
			Crv: was.JsonWebKeyCrv("Ed25519"),
			Kty: was.JsonWebKeyKty("OKP"),
		},
	}
}

func (b *MockJWKBuilder) WithKid(kid string) *MockJWKBuilder {
	b.jwk.Kid = kid
	return b
}

func (b *MockJWKBuilder) WithX(x string) *MockJWKBuilder {
	b.jwk.X = x
	return b
}

func (b *MockJWKBuilder) WithAlg(alg was.JsonWebKeyAlg) *MockJWKBuilder {
	b.jwk.Alg = alg
	return b
}

func (b *MockJWKBuilder) WithCrv(crv was.JsonWebKeyCrv) *MockJWKBuilder {
	b.jwk.Crv = crv
	return b
}

func (b *MockJWKBuilder) WithKty(kty was.JsonWebKeyKty) *MockJWKBuilder {
	b.jwk.Kty = kty
	return b
}

func (b *MockJWKBuilder) Build() was.JsonWebKey {
	return b.jwk
}

func (b *MockJWKBuilder) BuildSet() was.JsonWebKeySet {
	keys := []was.JsonWebKey{b.jwk}
	return was.JsonWebKeySet{
		Keys: &keys,
	}
}
