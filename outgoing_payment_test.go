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

func TestOutgoingPaymentGet(t *testing.T) {
	reqPath := "/outgoing-payments/456"
	mockResponse := testutils.NewMockOutgoingPaymentBuilder().Build()

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

	client.OutgoingPayment.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.OutgoingPayment.Get(context.Background(), openpayments.OutgoingPaymentGetParams{
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

func TestOutgoingPaymentList(t *testing.T) {
	reqPath := "/outgoing-payments"
	op := testutils.NewMockOutgoingPaymentBuilder().Build()

	mockResponse := openpayments.OutgoingPaymentListResponse{
		Pagination: rs.PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
		},
		Result: []rs.OutgoingPayment{op},
	}

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

	client.OutgoingPayment.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.OutgoingPayment.List(context.Background(), openpayments.OutgoingPaymentListParams{
		BaseURL:       mockServer.URL,
		AccessToken:   accessToken,
		WalletAddress: walletAddress,
	})

	assert.NoError(t, err)
	assert.Equal(t, &mockResponse, res)

	assert.Equal(t, 1, spy.CallCount())
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodGet, capture.Method)
	assert.Equal(t, strings.TrimPrefix(capture.Header.Get("Authorization"), "GNAP "), accessToken)
	assert.Equal(t, walletAddress, capture.URL.Query().Get("wallet-address"))
}

func TestOutgoingPaymentCreate(t *testing.T) {
	reqPath := "/outgoing-payments"
	mockResponse := testutils.NewMockOutgoingPaymentBuilder().Build()

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

	client.OutgoingPayment.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	payload := rs.CreateOutgoingPaymentRequest{}
	err = payload.FromCreateOutgoingPaymentWithQuote(rs.CreateOutgoingPaymentWithQuote{
		QuoteId:             "https://example.com/quotes/abc",
		WalletAddressSchema: walletAddress,
	})
	assert.NoError(t, err)

	res, err := client.OutgoingPayment.Create(context.Background(), openpayments.OutgoingPaymentCreateParams{
		BaseURL:     mockServer.URL,
		AccessToken: accessToken,
		Payload:     payload,
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
