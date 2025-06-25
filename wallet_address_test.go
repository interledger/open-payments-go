package openpayments_test

import (
	"context"
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
	defaultAuthServer := "https://auth.wallet.example/authorize"
	defaultID := "https://example.com/.well-known/pay"
	defaultResourceServer := "https://wallet.example/op"
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
	if err != nil {
		panic(err)
	}

	assert.NoError(t, err)
	assert.Equal(t, wa, res)
	assert.True(t, gock.IsDone())
}
