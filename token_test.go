package openpayments_test

import (
	"context"
	"log"
	"net/http"
	"strings"
	"testing"

	openpayments "github.com/interledger/open-payments-go"
	as "github.com/interledger/open-payments-go/generated/authserver"
	"github.com/interledger/open-payments-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestTokenRotate(t *testing.T) {
	reqPath := "/token/123"

	mockResponse := struct {
		AccessToken as.AccessToken `json:"access_token"`
	}{
		AccessToken: as.AccessToken{
			Value:  "new-access-token",
			Manage: "https://auth.example.com/token/456",
			Access: []as.AccessItem{},
		},
	}

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

	client.Token.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	res, err := client.Token.Rotate(context.Background(), openpayments.TokenRotateParams{
		URL:         mockServer.URL + reqPath,
		AccessToken: accessToken,
	})

	assert.NoError(t, err)
	assert.Equal(t, "new-access-token", res.Value)
	assert.Equal(t, "https://auth.example.com/token/456", res.Manage)

	assert.Equal(t, 1, spy.CallCount())
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodPost, capture.Method)
	assert.Equal(t, strings.TrimPrefix(capture.Header.Get("Authorization"), "GNAP "), accessToken)
	assert.Equal(t, mockServer.URL+reqPath, capture.URL.String())
}

func TestTokenRevoke(t *testing.T) {
	reqPath := "/token/123"

	mockServer := testutils.Mock(http.MethodDelete, reqPath, http.StatusNoContent, nil)
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

	client.Token.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	err = client.Token.Revoke(context.Background(), openpayments.TokenRevokeParams{
		URL:         mockServer.URL + reqPath,
		AccessToken: accessToken,
	})

	assert.NoError(t, err)

	assert.Equal(t, 1, spy.CallCount())
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodDelete, capture.Method)
	assert.Equal(t, strings.TrimPrefix(capture.Header.Get("Authorization"), "GNAP "), accessToken)
	assert.Equal(t, mockServer.URL+reqPath, capture.URL.String())

	assert.Equal(t, 1, spy.ResultCount())
	result := spy.Results[0]
	assert.Equal(t, http.StatusNoContent, result.Response.StatusCode)
}
