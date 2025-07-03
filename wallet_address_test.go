package openpayments_test

import (
	"context"
	"net/http"
	"testing"

	openpayments "github.com/interledger/open-payments-go"
	testutils "github.com/interledger/open-payments-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestWalletAddressGet(t *testing.T) {
	reqPath := "/.well-known/pay"
	wa := testutils.NewMockWalletAddressBuilder().Build()

	mockServer := testutils.Mock(http.MethodGet, reqPath, http.StatusOK, wa)
	defer mockServer.Close()
	client := openpayments.NewClient(openpayments.WithHTTPClientUnauthed(mockServer.Client()))

	res, err := client.WalletAddress.Get(context.Background(), openpayments.WalletAddressGetParams{
		URL: mockServer.URL + reqPath,
	})

	assert.NoError(t, err)
	assert.Equal(t, wa, res)
}

func TestWalletAddressKeysGet(t *testing.T) {
	reqPath := "/.well-known/pay"
	jwks := testutils.NewMockJWKBuilder().BuildSet()

	mockServer := testutils.Mock(http.MethodGet, reqPath+"/jwks.json", http.StatusOK, jwks)
	defer mockServer.Close()
	client := openpayments.NewClient(openpayments.WithHTTPClientUnauthed(mockServer.Client()))

	res, err := client.WalletAddress.GetKeys(context.Background(), openpayments.WalletAddressGetKeysParams{
		URL: mockServer.URL + reqPath,
	})

	assert.NoError(t, err)
	assert.Equal(t, jwks, res)
}
