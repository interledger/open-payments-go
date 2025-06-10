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
	"time"

	as "github.com/interledger/open-payments-go-sdk/pkg/generated/authserver"
	rs "github.com/interledger/open-payments-go-sdk/pkg/generated/resourceserver"
	schemas "github.com/interledger/open-payments-go-sdk/pkg/generated/schemas"
	op "github.com/interledger/open-payments-go-sdk/pkg/openpayments"
)

var (
	environment    *Environment
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
	walletAddress, err := unauthedClient.WalletAddress.Get(context.TODO(), op.WalletAddressGetParams{
		URL: url,
	})

	if err != nil {
		t.Fatalf("Error fetching wallet address: %v\n", err)
	}

	printJSON(t, walletAddress)
}

func TestUnauthedWalletAddressGetKeys(t *testing.T) {
	url := environment.ReceiverWalletAddressUrl
	t.Logf("\nunauthedClient.WalletAddress.GetKeys(\"%s\")\n", url)
	walletAddressKeys, err := unauthedClient.WalletAddress.GetKeys(context.TODO(), op.WalletAddressGetKeysParams{
		URL: url,
	})

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

	didDocument, err := unauthedClient.WalletAddress.GetDIDDocument(context.TODO(), op.WalletAddressGetDIDDocumentParams{
		URL: url,
	})
	if err != nil {
		t.Fatalf("Error fetching wallet address DID document: %v\n", err)
	}

	printJSON(t, didDocument)
}


func TestUnauthedGetPublicIncomingPayment(t *testing.T) {
	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}
	newIncomingPayment, err := newIncomingPayment(grant)
	if err != nil {
		t.Fatalf("Error creating new incoming payment: %v", err)
	}

	url, err := environment.RewriteURLIfNeeded(*newIncomingPayment.Id)
	if err != nil {
		t.Fatalf("Could not rewrite URL from incoming payment: %v", err)
	}

	t.Logf("\nunauthedClient.IncomingPayment.GetPublic(\"%s\")\n", url)

	incomingPayment, err := unauthedClient.IncomingPayment.GetPublic(context.TODO(), op.IncomingPaymentGetPublicParams{
		URL: url,
	})
	if err != nil {
		t.Fatalf("Error fetching incoming payment: %v\n", err)
	}

	printJSON(t, incomingPayment)
}

func TestGetPublicIncomingPayment(t *testing.T) {
	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}
	newIncomingPayment, err := newIncomingPayment(grant)
	if err != nil {
		t.Fatalf("Error creating new incoming payment: %v", err)
	}

	url := *newIncomingPayment.Id
	if environment.RewriteURL != nil {
		url, err = environment.RewriteURL(url)
		if err != nil {
			t.Fatalf("Could not rewrite URL from incoming payment: %v", err)
		}
	}

	t.Logf("\nAuthedClient.IncomingPayment.GetPublic(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.GetPublic(context.TODO(), op.IncomingPaymentGetPublicParams{
		URL: url,
	})
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
		op.GrantRequestParams{
			URL: environment.ReceiverOpenPaymentsAuthUrl,
			RequestBody: requestBody,
		},
	)
	if err != nil {
		t.Fatalf("Error with grant request: %v", err)
	}

	printJSON(t, grant)
}

func TestGrantCancel(t *testing.T){
	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error creating incoming payment grant: %v", err)
	}

	err = authedClient.Grant.Cancel(context.TODO(), op.GrantCancelParams{URL: grant.Continue.Uri, AccessToken: grant.Continue.AccessToken.Value})
  if err != nil {
		t.Errorf("Error canceling grant: %v", err)
	}

	err = authedClient.Grant.Cancel(context.TODO(), op.GrantCancelParams{URL: grant.Continue.Uri, AccessToken: grant.Continue.AccessToken.Value})

	if err == nil {
		t.Errorf("Grant cancellation did not error when expected")
	}
}

func TestGrantContinue(t *testing.T){
	t.Skip("Not implemented")
}

func TestAuthenticatedGetIncomingPayment(t *testing.T) {
	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}

	newIncomingPayment, err := newIncomingPayment(grant)
	if err != nil {
		t.Fatalf("Error creating new incoming payment: %v", err)
	}

	url := *newIncomingPayment.Id

	t.Logf("\nauthedClient.IncomingPayment.Get(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.Get(context.TODO(), op.IncomingPaymentGetParams{
		URL: url,
		AccessToken: grant.AccessToken.Value,
	})
	if err != nil {
		t.Fatalf("Error fetching incoming payment: %v", err)
	}

	printJSON(t, incomingPayment)
}

func TestListIncomingPayments(t *testing.T) {
	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}
	_, err = newIncomingPayment(grant)
	if err != nil {
		t.Fatalf("Error creating new incoming payment: %v", err)
	}

	url := environment.ReceiverOpenPaymentsResourceUrl

	t.Logf("\nauthedClient.IncomingPayment.List(\"%s\")\n", url)

	list, err := authedClient.IncomingPayment.List(context.TODO(), op.IncomingPaymentListParams{
		BaseURL: url,
		AccessToken: grant.AccessToken.Value,
		WalletAddress: environment.ResolvedReceiverWalletAddressUrl,
		Pagination: op.Pagination{
			First: "10",
		},
	})
	if err != nil {
		t.Fatalf("Error fetching incoming payments: %v", err)
	}

	printJSON(t, list)
}

func TestCreateIncomingPayment(t *testing.T) {
	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}

	url := environment.ReceiverOpenPaymentsResourceUrl
	t.Logf("\nauthedClient.IncomingPayment.Create(\"%s\")\n", url)

	expiresAt := time.Now().Add(24 * time.Hour)
	payload := rs.CreateIncomingPaymentJSONBody{
		WalletAddress: environment.ResolvedReceiverWalletAddressUrl,
		IncomingAmount: &schemas.Amount{
			Value: "100",
			AssetCode: environment.ReceiverAssetCode,
			AssetScale: environment.ReceiverAssetScale,
		},
		ExpiresAt: &expiresAt,
		Metadata: &map[string]interface{}{
			"description": "Free Money!",
		},
	}

	incomingPayment, err := authedClient.IncomingPayment.Create(context.TODO(), op.IncomingPaymentCreateParams{
		BaseURL: url,
		AccessToken: grant.AccessToken.Value,
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("Error creating incoming payment: %v", err)
	}

	printJSON(t, incomingPayment)

	if incomingPayment.Id == nil {
		t.Error("Expected incoming payment to have an ID")
	}
	if incomingPayment.Completed {
		t.Error("Expected new incoming payment to not be completed")
	}
	if incomingPayment.WalletAddress == nil || *incomingPayment.WalletAddress != environment.ResolvedReceiverWalletAddressUrl {
		t.Error("Expected wallet address to match request")
	}
}

func TestCompleteIncomingPayment(t *testing.T) {
	// Setup
	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}
	incomingPayment, err := newIncomingPayment(grant)
	if err != nil {
		t.Fatalf("Error creating incoming payment: %v", err)
	}
	if incomingPayment.Completed {
		t.Fatalf("Expected new incoming payment to not be completed")
	}

	// Complete the incoming payment
	t.Logf("\nauthedClient.IncomingPayment.Complete(\"%s\")\n", *incomingPayment.Id)

	completedPayment, err := authedClient.IncomingPayment.Complete(context.TODO(), op.IncomingPaymentCompleteParams{
		URL: *incomingPayment.Id,
		AccessToken: grant.AccessToken.Value,
	})
	if err != nil {
		t.Fatalf("Error completing incoming payment: %v", err)
	}

	printJSON(t, completedPayment)

	// Verify the payment is now completed
	if !completedPayment.Completed {
		t.Error("Expected completed incoming payment to be marked as completed")
	}
	if completedPayment.Id == nil || *completedPayment.Id != *incomingPayment.Id {
		t.Error("Expected completed payment ID to match original payment ID")
	}
}

func TestCreateAndGetQuote(t *testing.T) {
	// Setup
	incomingPaymentGrant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}
	newIncomingPayment, err := newIncomingPayment(incomingPaymentGrant)
	if err != nil {
		t.Fatalf("Error creating new incoming payment: %v", err)
	}
	if newIncomingPayment.Id == nil {
		t.Fatal("New incoming payment ID is nil")
	}

	quoteAccess := as.AccessQuote{
		Type: as.Quote,
		Actions: []as.AccessQuoteActions{
			// TODO: address how these arent scoped to quotes?
			// anti-corruption layer for the generated types?
			as.Create,
			as.Read,
		},
	}
	accessItem := as.AccessItem{}
	if err := accessItem.FromAccessQuote(quoteAccess); err != nil {
		t.Fatalf("Error creating AccessItem for quote: %v", err)
	}
	accessToken := struct {
		Access as.Access `json:"access"`
	}{
		Access: []as.AccessItem{accessItem},
	}
	quoteGrantRequestBody := as.PostRequestJSONBody{
		AccessToken: accessToken,
	}

	quoteGrant, err := authedClient.Grant.Request(
		context.TODO(),
		op.GrantRequestParams{
			URL:         environment.SenderOpenPaymentsAuthUrl,
			RequestBody: quoteGrantRequestBody,
		},
	)

	if err != nil {
		t.Fatalf("Error requesting grant for quote: %v", err)
	}
	if quoteGrant.AccessToken == nil || quoteGrant.AccessToken.Value == "" {
		t.Fatalf("Expected quote grant to have an access token")
	}

	// Create the Quote
	createQuotePayload := rs.CreateQuoteJSONBody0{
		WalletAddress: environment.ResolvedSenderWalletAddressUrl,
		Receiver:      *newIncomingPayment.Id,
		Method:        "ilp",
	}

	t.Logf("\nauthedClient.Quote.Create(\"%s\", %v)\n", environment.SenderOpenPaymentsResourceUrl, createQuotePayload)

	newQuote, err := authedClient.Quote.Create(context.TODO(), op.QuoteCreateParams{
		BaseURL:     environment.SenderOpenPaymentsResourceUrl,
		AccessToken: quoteGrant.AccessToken.Value,
		Payload:     createQuotePayload,
	})
	if err != nil {
		t.Fatalf("Error creating quote: %v", err)
	}

	printJSON(t, newQuote)

	if newQuote.Id == nil {
		t.Error("Expected new quote to have an ID")
	}
	if newQuote.WalletAddress == nil || *newQuote.WalletAddress != environment.ResolvedSenderWalletAddressUrl {
		t.Error("Expected quote wallet address to match sender's wallet address")
	}
	if newQuote.Receiver != *newIncomingPayment.Id {
		t.Errorf("Expected quote receiver to be %s, got %s", *newIncomingPayment.Id, newQuote.Receiver)
	}

	// Get the created Quote
	quoteURL, err := environment.RewriteURLIfNeeded(*newQuote.Id)
	if err != nil {
		t.Fatalf("Could not rewrite URL from quote: %v", err)
	}
	t.Logf("\nauthedClient.Quote.Get(\"%s\")\n", quoteURL)

	retrievedQuote, err := authedClient.Quote.Get(context.TODO(), op.QuoteGetParams{
		URL:         quoteURL,
		AccessToken: quoteGrant.AccessToken.Value,
	})
	if err != nil {
		t.Fatalf("Error fetching quote: %v", err)
	}

	printJSON(t, retrievedQuote)

	if retrievedQuote.Id == nil || *retrievedQuote.Id != *newQuote.Id {
		t.Errorf("Expected retrieved quote ID to be %s, got %s", *newQuote.Id, *retrievedQuote.Id)
	}
	if retrievedQuote.DebitAmount == (schemas.Amount{}) {
		t.Error("Expected retrieved quote to have a debitAmount")
	}
	if retrievedQuote.ReceiveAmount == (schemas.Amount{}) {
		t.Error("Expected retrieved quote to have a receiveAmount")
	}
}

// ==============
//  Test Helpers
// ==============
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
		op.GrantRequestParams{
			URL: environment.ReceiverOpenPaymentsAuthUrl,
			RequestBody: requestBody,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Error requesting grant: %w", err)
	}

	return &grant, nil
}

func newIncomingPayment(grant *op.Grant) (*rs.IncomingPaymentWithMethods, error) {
	url := environment.ReceiverOpenPaymentsResourceUrl
	expiresAt := time.Now().Add(24 * time.Hour)
	payload := rs.CreateIncomingPaymentJSONBody{
		WalletAddress: environment.ResolvedReceiverWalletAddressUrl,
		IncomingAmount: &schemas.Amount{
			Value: "100",
			AssetCode: environment.ReceiverAssetCode,
			AssetScale: environment.ReceiverAssetScale,
		},
		ExpiresAt: &expiresAt,
	}

	incomingPayment, err := authedClient.IncomingPayment.Create(context.TODO(), op.IncomingPaymentCreateParams{
		BaseURL: url,
		AccessToken: grant.AccessToken.Value,
		Payload: payload,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating incoming payment: %w", err)
	}

	return &incomingPayment, nil
}

func printJSON(t *testing.T, data interface{}) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Logf("%s\n", string(bytes))
}