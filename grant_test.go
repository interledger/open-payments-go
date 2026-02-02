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

var walletAddress = "https://example.com/.well-known/pay"
var pk = "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUVxZXptY1BoT0U4Ymt3TitqUXJwcGZSWXpHSWRGVFZXUUdUSEpJS3B6ODgKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo="
var keyID = "keyid"
var accessToken = "my-access-token"

func TestGrantCancel(t *testing.T) {
	reqPath := "/continue/1"

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

	client.Grant.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	err = client.Grant.Cancel(context.Background(), openpayments.GrantCancelParams{
		URL:         mockServer.URL + reqPath,
		AccessToken: accessToken,
	})

	assert.NoError(t, err)

	assert.Equal(t, spy.CallCount(), 1)
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodDelete, capture.Method)
	assert.Equal(t, strings.TrimPrefix(capture.Header.Get("Authorization"), "GNAP "), accessToken)
	assert.Equal(t, capture.URL.String(), mockServer.URL+reqPath)

	assert.Equal(t, spy.ResultCount(), 1)
	result := spy.Results[0]
	assert.Equal(t, result.Response.StatusCode, http.StatusNoContent)
}

func TestGrantRequest(t *testing.T) {
	reqPath := "/"

	mockResponse := openpayments.Grant{
		AccessToken: &as.AccessToken{
			Value:  "test-access-token",
			Manage: "https://auth.example.com/token/123",
			Access: []as.AccessItem{},
		},
		Continue: as.Continue{
			Uri: "https://auth.example.com/continue/123",
			AccessToken: struct {
				Value string `json:"value"`
			}{
				Value: "continue-token",
			},
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

	client.Grant.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	incomingAccess := as.AccessIncoming{
		Type: as.IncomingPayment,
		Actions: []as.AccessIncomingActions{
			as.AccessIncomingActionsCreate,
			as.AccessIncomingActionsRead,
		},
	}
	accessItem := as.AccessItem{}
	err = accessItem.FromAccessIncoming(incomingAccess)
	assert.NoError(t, err)

	grant, err := client.Grant.Request(context.Background(), openpayments.GrantRequestParams{
		URL: mockServer.URL + reqPath,
		RequestBody: as.GrantRequestWithAccessToken{
			AccessToken: as.AccessTokenRequest{
				Access: []as.AccessItem{accessItem},
			},
		},
	})

	assert.NoError(t, err)
	assert.True(t, grant.IsGranted())
	assert.Equal(t, "test-access-token", grant.AccessToken.Value)

	assert.Equal(t, 1, spy.CallCount())
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodPost, capture.Method)
	assert.Equal(t, "application/json", capture.Header.Get("Content-Type"))
	assert.Equal(t, mockServer.URL+reqPath, capture.URL.String())
}
