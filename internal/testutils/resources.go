package testutils

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	rs "github.com/interledger/open-payments-go/generated/resourceserver"
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
	rand.Read(bytes) // #nosec G104
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

// MockIncomingPaymentBuilder builds mock IncomingPaymentWithMethods for tests.
type MockIncomingPaymentBuilder struct {
	ip rs.IncomingPaymentWithMethods
}

func NewMockIncomingPaymentBuilder() *MockIncomingPaymentBuilder {
	defaultID := "https://example.com/incoming-payments/123"
	defaultWA := "https://example.com/.well-known/pay"
	return &MockIncomingPaymentBuilder{
		ip: rs.IncomingPaymentWithMethods{
			Id:            &defaultID,
			WalletAddress: &defaultWA,
			Completed:     false,
			CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			ReceivedAmount: rs.Amount{
				Value:      "0",
				AssetCode:  "USD",
				AssetScale: 2,
			},
			Methods: []rs.IncomingPaymentWithMethods_Methods_Item{},
		},
	}
}

func (b *MockIncomingPaymentBuilder) WithID(id string) *MockIncomingPaymentBuilder {
	b.ip.Id = &id
	return b
}

func (b *MockIncomingPaymentBuilder) WithWalletAddress(wa string) *MockIncomingPaymentBuilder {
	b.ip.WalletAddress = &wa
	return b
}

func (b *MockIncomingPaymentBuilder) WithCompleted(completed bool) *MockIncomingPaymentBuilder {
	b.ip.Completed = completed
	return b
}

func (b *MockIncomingPaymentBuilder) WithIncomingAmount(amount rs.Amount) *MockIncomingPaymentBuilder {
	b.ip.IncomingAmount = &amount
	return b
}

func (b *MockIncomingPaymentBuilder) WithReceivedAmount(amount rs.Amount) *MockIncomingPaymentBuilder {
	b.ip.ReceivedAmount = amount
	return b
}

func (b *MockIncomingPaymentBuilder) Build() rs.IncomingPaymentWithMethods {
	return b.ip
}

// MockOutgoingPaymentBuilder builds mock OutgoingPayment for tests.
type MockOutgoingPaymentBuilder struct {
	op rs.OutgoingPayment
}

func NewMockOutgoingPaymentBuilder() *MockOutgoingPaymentBuilder {
	defaultID := "https://example.com/outgoing-payments/456"
	defaultWA := "https://example.com/.well-known/pay"
	defaultQuoteID := "https://example.com/quotes/abc"
	return &MockOutgoingPaymentBuilder{
		op: rs.OutgoingPayment{
			Id:            &defaultID,
			WalletAddress: &defaultWA,
			CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Receiver:      "https://receiver.example.com/incoming-payments/789",
			QuoteId:       &defaultQuoteID,
			DebitAmount: rs.Amount{
				Value:      "1000",
				AssetCode:  "USD",
				AssetScale: 2,
			},
			ReceiveAmount: rs.Amount{
				Value:      "1000",
				AssetCode:  "USD",
				AssetScale: 2,
			},
			SentAmount: rs.Amount{
				Value:      "0",
				AssetCode:  "USD",
				AssetScale: 2,
			},
		},
	}
}

func (b *MockOutgoingPaymentBuilder) WithID(id string) *MockOutgoingPaymentBuilder {
	b.op.Id = &id
	return b
}

func (b *MockOutgoingPaymentBuilder) WithWalletAddress(wa string) *MockOutgoingPaymentBuilder {
	b.op.WalletAddress = &wa
	return b
}

func (b *MockOutgoingPaymentBuilder) WithReceiver(receiver string) *MockOutgoingPaymentBuilder {
	b.op.Receiver = receiver
	return b
}

func (b *MockOutgoingPaymentBuilder) WithQuoteID(quoteID string) *MockOutgoingPaymentBuilder {
	b.op.QuoteId = &quoteID
	return b
}

func (b *MockOutgoingPaymentBuilder) Build() rs.OutgoingPayment {
	return b.op
}

// MockQuoteBuilder builds mock Quote for tests.
type MockQuoteBuilder struct {
	q rs.Quote
}

func NewMockQuoteBuilder() *MockQuoteBuilder {
	defaultID := "https://example.com/quotes/abc"
	defaultWA := "https://example.com/.well-known/pay"
	return &MockQuoteBuilder{
		q: rs.Quote{
			Id:            &defaultID,
			WalletAddress: &defaultWA,
			CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Receiver:      "https://receiver.example.com/incoming-payments/789",
			Method:        rs.PaymentMethodIlp,
			DebitAmount: rs.Amount{
				Value:      "1000",
				AssetCode:  "USD",
				AssetScale: 2,
			},
			ReceiveAmount: rs.Amount{
				Value:      "1000",
				AssetCode:  "USD",
				AssetScale: 2,
			},
		},
	}
}

func (b *MockQuoteBuilder) WithID(id string) *MockQuoteBuilder {
	b.q.Id = &id
	return b
}

func (b *MockQuoteBuilder) WithWalletAddress(wa string) *MockQuoteBuilder {
	b.q.WalletAddress = &wa
	return b
}

func (b *MockQuoteBuilder) WithReceiver(receiver string) *MockQuoteBuilder {
	b.q.Receiver = receiver
	return b
}

func (b *MockQuoteBuilder) Build() rs.Quote {
	return b.q
}
