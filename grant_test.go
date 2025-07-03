package openpayments_test

import (
	"context"
	"net/http"
	"testing"

	openpayments "github.com/interledger/open-payments-go"
	"github.com/interledger/open-payments-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var pk string
var client *openpayments.AuthenticatedClient

func init() {
	pk = "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUVxZXptY1BoT0U4Ymt3TitqUXJwcGZSWXpHSWRGVFZXUUdUSEpJS3B6ODgKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo="
	client = openpayments.NewAuthenticatedClient("https://example.com/.well-known/pay", pk, "random")
}

func TestGrantCancel(t *testing.T) {
	var capture *http.Request

	spy := testutils.Spy(http.StatusNoContent, &capture)

	client.Grant.DoSigned = spy

	err := client.Grant.Cancel(context.Background(), openpayments.GrantCancelParams{
		URL:         "https://test.com",
		AccessToken: "myaccesstoken",
	})
	assert.NoError(t, err)
	assert.NotNil(t, capture)
	assert.Equal(t, capture.Method, http.MethodDelete)
	assert.Equal(t, capture.URL.String(), "https://test.com")
	assert.Equal(t, capture.Header.Get("authorization"), "GNAP myaccesstoken")
}
