package openpayments_test

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
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

	var requestBody as.GrantRequest
	err = requestBody.FromGrantRequestWithAccessToken(as.GrantRequestWithAccessToken{
		AccessToken: as.AccessTokenRequest{
			Access: []as.AccessItem{accessItem},
		},
	})
	assert.NoError(t, err)

	grant, err := client.Grant.Request(context.Background(), openpayments.GrantRequestParams{
		URL:         mockServer.URL + reqPath,
		RequestBody: requestBody,
	})

	assert.NoError(t, err)
	assert.True(t, grant.IsGrantedWithAccessToken())
	assert.Equal(t, "test-access-token", grant.AccessToken.Value)

	assert.Equal(t, 1, spy.CallCount())
	capture := spy.Calls[0]
	assert.Equal(t, http.MethodPost, capture.Method)
	assert.Equal(t, "application/json", capture.Header.Get("Content-Type"))
	assert.Equal(t, mockServer.URL+reqPath, capture.URL.String())
}

func TestGrantGrantedStatus(t *testing.T) {
	tests := []struct {
		name                string
		grant               openpayments.Grant
		wantWithAccessToken bool
		wantWithSubject     bool
	}{
		{
			name:                "access token grant",
			grant:               openpayments.Grant{AccessToken: &as.AccessToken{}},
			wantWithAccessToken: true,
		},
		{
			name:            "subject grant",
			grant:           openpayments.Grant{Subject: &as.Subject{}},
			wantWithSubject: true,
		},
		{
			name:  "pending grant",
			grant: openpayments.Grant{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantWithAccessToken, tt.grant.IsGrantedWithAccessToken())
			assert.Equal(t, tt.wantWithSubject, tt.grant.IsGrantedWithSubject())
		})
	}
}

func TestGrantRequest_WithSubject(t *testing.T) {
	var receivedBody []byte
	mockResponse := openpayments.Grant{
		Interact: &as.InteractResponse{
			Redirect: "https://auth.example.com/interact/abc",
			Finish:   "finish-nonce",
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(server.Client()))
	if err != nil {
		log.Fatalf("Failed to initialize authenticated client: %v", err)
	}

	subject := as.Subject{
		SubIds: []struct {
			Format as.SubjectSubIdsFormat `json:"format"`
			Id     string                 `json:"id"`
		}{{Format: "uri", Id: "https://example.com/alice"}},
	}
	interact := as.InteractRequest{
		Start: []as.InteractRequestStart{as.InteractRequestStartRedirect},
	}

	var requestBody as.GrantRequest
	err = requestBody.FromGrantRequestWithSubject(as.GrantRequestWithSubject{
		Subject:  subject,
		Interact: interact,
	})
	assert.NoError(t, err)

	grant, err := client.Grant.Request(context.Background(), openpayments.GrantRequestParams{
		URL:         server.URL + "/",
		RequestBody: requestBody,
	})
	assert.NoError(t, err)
	assert.True(t, grant.IsInteractive())

	var sent map[string]any
	assert.NoError(t, json.Unmarshal(receivedBody, &sent))
	assert.NotNil(t, sent["subject"], "subject variant should be preserved after setClient")
	assert.NotNil(t, sent["interact"], "interact should be preserved (required for subject grants)")
	clientObj, ok := sent["client"].(map[string]any)
	assert.True(t, ok, "client should be injected as an object")
	assert.Equal(t, walletAddress, clientObj["walletAddress"])
}

func TestGrantRequest_WithAccessTokenAndSubject(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(openpayments.Grant{
			AccessToken: &as.AccessToken{},
			Continue:    as.Continue{},
		})
	}))
	defer server.Close()

	client, err := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(server.Client()))
	assert.NoError(t, err)

	accessItem := as.AccessItem{}
	assert.NoError(t, accessItem.FromAccessIncoming(as.AccessIncoming{
		Type:    as.IncomingPayment,
		Actions: []as.AccessIncomingActions{as.AccessIncomingActionsCreate},
	}))

	var requestBody as.GrantRequest
	assert.NoError(t, requestBody.FromGrantRequestWithAccessToken(as.GrantRequestWithAccessToken{
		AccessToken: as.AccessTokenRequest{Access: []as.AccessItem{accessItem}},
		Subject:     &as.Subject{},
	}))

	_, err = client.Grant.Request(context.Background(), openpayments.GrantRequestParams{
		URL:         server.URL,
		RequestBody: requestBody,
	})
	assert.NoError(t, err)

	var sent map[string]any
	assert.NoError(t, json.Unmarshal(receivedBody, &sent))
	assert.NotNil(t, sent["access_token"])
	assert.NotNil(t, sent["subject"])
	assert.NotContains(t, sent, "interact", "access-token variant should be preserved")
}

func TestGrantRequest_WithClientOverride(t *testing.T) {
	var receivedBody []byte
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(server.Client()))
	if err != nil {
		log.Fatalf("Failed to initialize authenticated client: %v", err)
	}

	incomingAccess := as.AccessIncoming{
		Type:    as.IncomingPayment,
		Actions: []as.AccessIncomingActions{as.AccessIncomingActionsCreate},
	}
	accessItem := as.AccessItem{}
	assert.NoError(t, accessItem.FromAccessIncoming(incomingAccess))

	var requestBody as.GrantRequest
	err = requestBody.FromGrantRequestWithAccessToken(as.GrantRequestWithAccessToken{
		AccessToken: as.AccessTokenRequest{
			Access: []as.AccessItem{accessItem},
		},
	})
	assert.NoError(t, err)

	jwk := as.JsonWebKey{
		Kid: "key1",
		Alg: "EdDSA",
		Kty: "OKP",
		Crv: "Ed25519",
		X:   "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo",
	}

	_, err = client.Grant.Request(context.Background(), openpayments.GrantRequestParams{
		URL:            server.URL + "/",
		RequestBody:    requestBody,
		ClientOverride: &as.ClientDirectedIdentity{Jwk: jwk},
	})
	assert.NoError(t, err)

	var sent map[string]any
	assert.NoError(t, json.Unmarshal(receivedBody, &sent))
	clientObj, ok := sent["client"].(map[string]any)
	assert.True(t, ok, "client should be an object")
	assert.NotContains(t, clientObj, "walletAddress", "override should replace wallet address")
	sentJwk, ok := clientObj["jwk"].(map[string]any)
	assert.True(t, ok, "client.jwk should be present")
	assert.Equal(t, "key1", sentJwk["kid"])
	assert.Equal(t, "EdDSA", sentJwk["alg"])
}

func TestGrantRequest_WithCardAuthorization(t *testing.T) {
	pinBlock := "23e412341234"

	tests := []struct {
		name              string
		cardAuthorization as.CardAuthorization
	}{
		{
			name: "Card Authorization with PIN",
			cardAuthorization: as.CardAuthorization{
				TlvData:  "9F26089F27019F1002",
				PinBlock: &pinBlock,
				Pwk: &as.PinWorkingKey{
					Key: "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4",
					Kvc: "A3F1C2",
					Iv:  "00000000000000000000000000000000",
				},
				RequestId: "D01A53B6-9752-490C-9865-42753E186823",
			},
		},
		{
			name: "Card Authorization without PIN",
			cardAuthorization: as.CardAuthorization{
				TlvData:   "9F26089F27019F1002",
				RequestId: "9148BD21-863C-4C4B-9E8F-24F01EB6B3DD",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestAccessItem := as.AccessItem{}
			err := requestAccessItem.FromAccessOutgoing(as.AccessOutgoing{
				Type:              as.OutgoingPayment,
				Actions:           []as.AccessOutgoingActions{as.AccessOutgoingActionsCreate, as.AccessOutgoingActionsRead},
				Identifier:        walletAddress,
				CardAuthorization: &tt.cardAuthorization,
			})
			assert.NoError(t, err)

			var requestBody as.GrantRequest
			err = requestBody.FromGrantRequestWithAccessToken(as.GrantRequestWithAccessToken{
				AccessToken: as.AccessTokenRequest{Access: []as.AccessItem{requestAccessItem}},
			})
			assert.NoError(t, err)

			responseCardAuthorization := &as.CardAuthorization{
				TlvData:   "8A023030...",
				RequestId: tt.cardAuthorization.RequestId,
			}
			responseAccessItem := as.AccessItem{}
			err = responseAccessItem.FromAccessOutgoing(as.AccessOutgoing{
				Type:              as.OutgoingPayment,
				Actions:           []as.AccessOutgoingActions{as.AccessOutgoingActionsCreate, as.AccessOutgoingActionsRead},
				Identifier:        walletAddress,
				CardAuthorization: responseCardAuthorization,
			})
			assert.NoError(t, err)

			mockResponse := openpayments.Grant{
				AccessToken: &as.AccessToken{
					Value:  "test-access-token",
					Manage: "https://auth.example.com/token/123",
					Access: []as.AccessItem{responseAccessItem},
				},
				Continue: as.Continue{},
			}

			var receivedBody []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedBody, _ = io.ReadAll(r.Body)
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(mockResponse)
			}))
			defer server.Close()

			client, err := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(server.Client()))
			assert.NoError(t, err)

			grant, err := client.Grant.Request(context.Background(), openpayments.GrantRequestParams{
				URL:         server.URL + "/",
				RequestBody: requestBody,
			})
			assert.NoError(t, err)

			var sentRequest as.GrantRequest
			assert.NoError(t, json.Unmarshal(receivedBody, &sentRequest))
			sentTokenRequest, err := sentRequest.AsGrantRequestWithAccessToken()
			assert.NoError(t, err)
			assert.Len(t, sentTokenRequest.AccessToken.Access, 1)
			sentOutgoingAccess, err := sentTokenRequest.AccessToken.Access[0].AsAccessOutgoing()
			assert.NoError(t, err)
			assert.Equal(t, &tt.cardAuthorization, sentOutgoingAccess.CardAuthorization)

			assert.NotNil(t, grant.AccessToken)
			assert.Len(t, grant.AccessToken.Access, 1)
			grantedOutgoingAccess, err := grant.AccessToken.Access[0].AsAccessOutgoing()
			assert.NoError(t, err)
			assert.Equal(t, responseCardAuthorization, grantedOutgoingAccess.CardAuthorization)
		})
	}
}
