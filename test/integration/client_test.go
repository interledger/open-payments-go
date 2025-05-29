package integration

/*
Integration tests for the Open Payments SDK. Requires a running instance of Rafiki.

The tests using the authenticated client unless the test names specifically denote
otherwise (e.g. TestUnauthedWalletAddressGet).
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	as "github.com/interledger/open-payments-go-sdk/pkg/generated/authserver"
	op "github.com/interledger/open-payments-go-sdk/pkg/openpayments"
)

var (
	environment    Environment
	unauthedClient *op.Client
	authedClient *op.AuthenticatedClient
)

func TestMain(m *testing.M) {
	// TODO: switch on some ENV var/cli arg when NewTestnetEnvironment is 
	// implemented to get correct environment
	environment = NewLocalEnvironment()
	
	// TODO: have NewLocalEnvironment() just return the initialized clients?
	unauthedClient = op.NewClient(op.WithHTTPClientUnauthed(environment.HttpClient))
	authedClient = op.NewAuthenticatedClient(
		environment.ClientWalletAddressURL,
		environment.PrivateKey,
		environment.KeyId,
		op.WithHTTPClientAuthed(environment.HttpClient),
		op.WithPreSignHook(environment.PreSignHook),
		op.WithPostSignHook(environment.PostSignHook),
	)

	os.Exit(m.Run())
}


// TODO: test wa methods on authed client.
// - could combine with current unauthed (just TestWalletAddressget??) but
//   maybe seperate is better?

func TestUnauthedWalletAddressGet(t *testing.T) {
	url := environment.ReceiverWalletAddressUrl
	t.Logf("\nunauthedClient.WalletAddress.Get(\"%s\")\n", url)
	walletAddress, err := unauthedClient.WalletAddress.Get(context.TODO(), url)

	if err != nil {
		t.Fatalf("Error fetching wallet address: %v\n", err)
	}

	printJSON(t, walletAddress)
}

func TestUnauthedWalletAddressGetKeys(t *testing.T) {
	url := environment.ReceiverWalletAddressUrl
	t.Logf("\nunauthedClient.WalletAddress.GetKeys(\"%s\")\n", url)
	walletAddressKeys, err := unauthedClient.WalletAddress.GetKeys(context.TODO(), url)

	if err != nil {
		t.Fatalf("Error fetching wallet address keys: %v\n", err)
		return
	}

	printJSON(t, walletAddressKeys)
}

func TestUnauthedWalletAddressGetDIDDocument(t *testing.T) {
	t.Skip("Skipping: DID Document endpoint not implemented in Rafiki yet")

	url := environment.ReceiverWalletAddressUrl
	t.Logf("\nunauthedClient.WalletAddress.GetDIDDocument(\"%s\")\n", url)

	didDocument, err := unauthedClient.WalletAddress.GetDIDDocument(context.TODO(), url)
	if err != nil {
		t.Fatalf("Error fetching wallet address DID document: %v\n", err)
	}

	printJSON(t, didDocument)
}


// TODO: setup an incoming payment to get properly
func TestUnauthedGetPublicIncomingPayment(t *testing.T) {
	incomingPaymentId := "67684fda-f33d-4ce0-aaae-38376b40311c"
	url := fmt.Sprintf("http://localhost:4000/incoming-payments/%s", incomingPaymentId)

	t.Logf("\nunauthedClient.IncomingPayment.GetPublic(\"%s\")\n", url)

	incomingPayment, err := unauthedClient.IncomingPayment.GetPublic(context.TODO(), url)
	if err != nil {
		t.Fatalf("Error fetching incoming payment: %v\n", err)
	}

	printJSON(t, incomingPayment)
}

// TODO: setup an incoming payment to get properly
func TestGetPublicIncomingPayment(t *testing.T) {
	incomingPaymentId := "67684fda-f33d-4ce0-aaae-38376b40311c"
	url := fmt.Sprintf("http://localhost:4000/incoming-payments/%s", incomingPaymentId)

	t.Logf("\nAuthedClient.IncomingPayment.GetPublic(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.GetPublic(context.TODO(), url)
	if err != nil {
		t.Fatalf("Error fetching incoming payment: %v\n", err)
	}

	printJSON(t, incomingPayment)
}

func TestGrantRequestIncomingPayment(t *testing.T) {
	incomingAccess := as.AccessIncoming{
		Type: as.IncomingPayment,
		Actions: []as.AccessIncomingActions{
			as.AccessIncomingActionsCreate,
			as.AccessIncomingActionsRead,
			as.AccessIncomingActionsList,
			as.AccessIncomingActionsComplete,
		},
	}
	accessItem := as.AccessItem{}
	if err := accessItem.FromAccessIncoming(incomingAccess); err != nil {
		t.Fatalf("Error creating AccessItem: %v", err)
	}
	accessToken := struct {
		Access as.Access `json:"access"`
	}{
		Access: []as.AccessItem{accessItem},
	}

	requestBody := as.PostRequestJSONBody{
		AccessToken: accessToken,
	}

	grant, err := authedClient.Grant.Request(
		context.TODO(),
		environment.ReceiverOpenPaymentsAuthUrl,
		requestBody,
	)
	if err != nil {
		t.Fatalf("Error with grant request: %v", err)
	}

	printJSON(t, grant)
}


// TODO: 
// - setup an incoming payment to get properly
func TestAuthenticatedGetIncomingPayment(t *testing.T) {
	// t.Skip("Skipping: Not fully implemented")

	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}

	incomingPaymentId := "67684fda-f33d-4ce0-aaae-38376b40311c"
	url := fmt.Sprintf("%s/incoming-payments/%s", environment.ReceiverOpenPaymentsResourceUrl, incomingPaymentId)

	t.Logf("\nauthedClient.IncomingPayment.Get(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.Get(context.TODO(), url, grant.AccessToken.Value)
	if err != nil {
		t.Fatalf("Error fetching incoming payment: %v", err)
	}

	printJSON(t, incomingPayment)
	
	// t.Fatal("not fully implemented")
}


// TODO: 
// - setup an incoming payment to get properly
func TestListIncomingPayments(t *testing.T) {
	t.Skip("Skipping: Not fully implemented")

	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}

	url := fmt.Sprintf("%s/incoming-payments/", environment.ReceiverOpenPaymentsResourceUrl)

	t.Logf("\nauthedClient.IncomingPayment.List(\"%s\")\n", url)

	list, err := authedClient.IncomingPayment.List(context.TODO(), url, grant.AccessToken.Value, op.ListArgs{
		WalletAddress: environment.ReceiverWalletAddressUrl,
		Pagination: op.Pagination{
			First:  "10",
			// Cursor: "abc123",
		},
	})
	if err != nil {
		t.Fatalf("Error fetching incoming payments: %v", err)
	}

	printJSON(t, list)

	t.Fatal("not fully implemented")
}

func newIncomingPaymentGrant() (*op.Grant, error) {
		incomingAccess := as.AccessIncoming{
		Type: as.IncomingPayment,
		Actions: []as.AccessIncomingActions{
			as.AccessIncomingActionsCreate,
			as.AccessIncomingActionsRead,
			as.AccessIncomingActionsList,
			as.AccessIncomingActionsComplete,
		},
	}
	accessItem := as.AccessItem{}
	if err := accessItem.FromAccessIncoming(incomingAccess); err != nil {
		return nil, fmt.Errorf("Error creating AccessItem: %w", err)
	}
	accessToken := struct {
		Access as.Access `json:"access"`
	}{
		Access: []as.AccessItem{accessItem},
	}

	requestBody := as.PostRequestJSONBody{
		AccessToken: accessToken,
	}

	grant, err := authedClient.Grant.Request(
		context.TODO(),
		environment.ReceiverOpenPaymentsAuthUrl,
		requestBody,
	)

	if err != nil {
		return nil, fmt.Errorf("Error requesting grant: %w", err)
	}

	return &grant, nil
}

func printJSON(t *testing.T, data interface{}) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("%s\n", string(bytes))
}