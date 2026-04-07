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

func TestNewIncomingPaymentGrantRequestParams(t *testing.T) {
	url := "https://auth.example.com"
	actions := []as.AccessIncomingActions{
		as.AccessIncomingActionsCreate,
		as.AccessIncomingActionsRead,
	}

	params, err := openpayments.NewIncomingPaymentGrantRequestParams(url, actions)
	assert.NoError(t, err)
	assert.Equal(t, url, params.URL)
	assert.Len(t, params.RequestBody.AccessToken.Access, 1)

	// Verify the access item was constructed correctly by round-tripping
	incoming, err := params.RequestBody.AccessToken.Access[0].AsAccessIncoming()
	assert.NoError(t, err)
	assert.Equal(t, as.IncomingPayment, incoming.Type)
	assert.Equal(t, actions, incoming.Actions)
}

func TestNewQuoteGrantRequestParams(t *testing.T) {
	url := "https://auth.example.com"
	actions := []as.AccessQuoteActions{as.Create, as.Read}

	params, err := openpayments.NewQuoteGrantRequestParams(url, actions)
	assert.NoError(t, err)
	assert.Equal(t, url, params.URL)
	assert.Len(t, params.RequestBody.AccessToken.Access, 1)

	quote, err := params.RequestBody.AccessToken.Access[0].AsAccessQuote()
	assert.NoError(t, err)
	assert.Equal(t, as.Quote, quote.Type)
	assert.Equal(t, actions, quote.Actions)
}

func TestNewOutgoingPaymentGrantRequestParams(t *testing.T) {
	url := "https://auth.example.com"
	identifier := "https://wallet.example.com/.well-known/pay"
	actions := []as.AccessOutgoingActions{
		as.AccessOutgoingActionsCreate,
		as.AccessOutgoingActionsRead,
	}

	params, err := openpayments.NewOutgoingPaymentGrantRequestParams(url, identifier, actions)
	assert.NoError(t, err)
	assert.Equal(t, url, params.URL)
	assert.Len(t, params.RequestBody.AccessToken.Access, 1)

	outgoing, err := params.RequestBody.AccessToken.Access[0].AsAccessOutgoing()
	assert.NoError(t, err)
	assert.Equal(t, as.OutgoingPayment, outgoing.Type)
	assert.Equal(t, identifier, outgoing.Identifier)
	assert.Equal(t, actions, outgoing.Actions)
}

func TestGrantRequestOptionWithInteract(t *testing.T) {
	interact := &as.InteractRequest{
		Start: []as.InteractRequestStart{as.InteractRequestStartRedirect},
	}

	params, err := openpayments.NewIncomingPaymentGrantRequestParams(
		"https://auth.example.com",
		[]as.AccessIncomingActions{as.AccessIncomingActionsCreate},
		openpayments.WithInteract(interact),
	)
	assert.NoError(t, err)
	assert.Equal(t, interact, params.RequestBody.Interact)
}

func TestGrantRequestOptionWithSubject(t *testing.T) {
	subject := &as.Subject{}

	params, err := openpayments.NewIncomingPaymentGrantRequestParams(
		"https://auth.example.com",
		[]as.AccessIncomingActions{as.AccessIncomingActionsCreate},
		openpayments.WithSubject(subject),
	)
	assert.NoError(t, err)
	assert.Equal(t, subject, params.RequestBody.Subject)
}
