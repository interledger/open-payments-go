package openpayments_test

import (
	"context"
	"log"
	"net/http"
	"strings"
	"testing"

	openpayments "github.com/interledger/open-payments-go"
	rs "github.com/interledger/open-payments-go/generated/resourceserver"
	"github.com/interledger/open-payments-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestQuoteGet(t *testing.T) {
	reqPath := "/quotes/abc"
	mockResponse := testutils.NewMockQuoteBuilder().Build()

	mockServer := testutils.Mock(http.MethodGet, reqPath, http.StatusOK, mockResponse)
	defer mockServer.Close()

	client, err := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(mockServer.Client()))
	if err != nil {
		log.Fatalf("Failed to initialize authenticated client: %v", err)
	}

	ds := func(req *http.Request) testutils.DoSignedResult {
		res, err := client.DoSigned(req)
		return testutils.DoSignedResult{Response: res, Error: err}
	}
	spy := testutils.SpyOn(ds)

	client.Quote.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.Quote.Get(context.Background(), openpayments.QuoteGetParams{
		URL:         mockServer.URL + reqPath,
		AccessToken: accessToken,
	})

	assert.NoError(t, err)
	assert.Equal(t, mockResponse, res)

	assert.Equal(t, 1, spy.CallCount())
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodGet, capture.Method)
	assert.Equal(t, strings.TrimPrefix(capture.Header.Get("Authorization"), "GNAP "), accessToken)
	assert.Equal(t, mockServer.URL+reqPath, capture.URL.String())
}

func TestQuoteCreate(t *testing.T) {
	reqPath := "/quotes"
	mockResponse := testutils.NewMockQuoteBuilder().Build()

	mockServer := testutils.Mock(http.MethodPost, reqPath, http.StatusCreated, mockResponse)
	defer mockServer.Close()

	client, err := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(mockServer.Client()))
	if err != nil {
		log.Fatalf("Failed to initialize authenticated client: %v", err)
	}

	ds := func(req *http.Request) testutils.DoSignedResult {
		res, err := client.DoSigned(req)
		return testutils.DoSignedResult{Response: res, Error: err}
	}
	spy := testutils.SpyOn(ds)

	client.Quote.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.Quote.Create(context.Background(), openpayments.QuoteCreateParams{
		BaseURL:     mockServer.URL,
		AccessToken: accessToken,
		Payload: rs.CreateQuoteJSONBody0{
			Method:              rs.PaymentMethodIlp,
			Receiver:            "https://receiver.example.com/incoming-payments/789",
			WalletAddressSchema: walletAddress,
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, mockResponse, res)

	assert.Equal(t, 1, spy.CallCount())
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodPost, capture.Method)
	assert.Equal(t, "application/json", capture.Header.Get("Content-Type"))
	assert.Equal(t, strings.TrimPrefix(capture.Header.Get("Authorization"), "GNAP "), accessToken)
	assert.Equal(t, mockServer.URL+reqPath, capture.URL.String())
}
