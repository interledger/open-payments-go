package openpayments_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	openpayments "github.com/interledger/open-payments-go"
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

	client := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(mockServer.Client()))

	ds := func(req *http.Request) testutils.DoSignedResult {
		res, err := client.DoSigned(req)
		return testutils.DoSignedResult{Response: res, Error: err}
	}
	spy := testutils.SpyOn(ds)

	client.Grant.DoSigned = func(req *http.Request) (*http.Response, error) {
		result := spy.Func()(req)
		return result.Response, result.Error
	}

	err := client.Grant.Cancel(context.Background(), openpayments.GrantCancelParams{
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
