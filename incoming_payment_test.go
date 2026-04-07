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

func TestIncomingPaymentGetPublic(t *testing.T) {
	reqPath := "/incoming-payments/123"

	mockResponse := rs.PublicIncomingPayment{
		AuthServer: "https://auth.example.com",
		ReceivedAmount: &rs.Amount{
			Value:      "500",
			AssetCode:  "USD",
			AssetScale: 2,
		},
	}

	mockServer := testutils.Mock(http.MethodGet, reqPath, http.StatusOK, mockResponse)
	defer mockServer.Close()

	client := openpayments.NewClient(openpayments.WithHTTPClientUnauthed(mockServer.Client()))

	res, err := client.IncomingPayment.GetPublic(context.Background(), openpayments.IncomingPaymentGetPublicParams{
		URL: mockServer.URL + reqPath,
	})

	assert.NoError(t, err)
	assert.Equal(t, mockResponse, res)
}

func TestIncomingPaymentGet(t *testing.T) {
	reqPath := "/incoming-payments/123"
	mockResponse := testutils.NewMockIncomingPaymentBuilder().Build()

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

	client.IncomingPayment.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.IncomingPayment.Get(context.Background(), openpayments.IncomingPaymentGetParams{
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

func TestIncomingPaymentList(t *testing.T) {
	reqPath := "/incoming-payments"
	ip := testutils.NewMockIncomingPaymentBuilder().Build()

	mockResponse := openpayments.IncomingPaymentListResponse{
		Pagination: rs.PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
		},
		Result: []rs.IncomingPaymentWithMethods{ip},
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

	client.IncomingPayment.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.IncomingPayment.List(context.Background(), openpayments.IncomingPaymentListParams{
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

func TestIncomingPaymentCreate(t *testing.T) {
	reqPath := "/incoming-payments"
	mockResponse := testutils.NewMockIncomingPaymentBuilder().
		WithIncomingAmount(rs.Amount{Value: "1000", AssetCode: "USD", AssetScale: 2}).
		Build()

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

	client.IncomingPayment.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.IncomingPayment.Create(context.Background(), openpayments.IncomingPaymentCreateParams{
		BaseURL:     mockServer.URL,
		AccessToken: accessToken,
		Payload: rs.CreateIncomingPaymentJSONBody{
			WalletAddressSchema: walletAddress,
			IncomingAmount: &rs.Amount{
				Value:      "1000",
				AssetCode:  "USD",
				AssetScale: 2,
			},
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

func TestIncomingPaymentComplete(t *testing.T) {
	reqPath := "/incoming-payments/123/complete"
	mockResponse := testutils.NewMockIncomingPaymentBuilder().
		WithCompleted(true).
		Build()

	mockServer := testutils.Mock(http.MethodPost, reqPath, http.StatusOK, mockResponse)
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

	client.IncomingPayment.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.IncomingPayment.Complete(context.Background(), openpayments.IncomingPaymentCompleteParams{
		URL:         mockServer.URL + "/incoming-payments/123",
		AccessToken: accessToken,
	})

	assert.NoError(t, err)
	assert.True(t, res.Completed)
	assert.Equal(t, mockResponse, res)

	assert.Equal(t, 1, spy.CallCount())
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodPost, capture.Method)
	assert.Equal(t, strings.TrimPrefix(capture.Header.Get("Authorization"), "GNAP "), accessToken)
	assert.Equal(t, mockServer.URL+reqPath, capture.URL.String())
}
