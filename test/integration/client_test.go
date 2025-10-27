package integration

/*
Integration tests for the Open Payments SDK. Requires a running instance of Rafiki.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	op "github.com/interledger/open-payments-go"
	as "github.com/interledger/open-payments-go/generated/authserver"
	rs "github.com/interledger/open-payments-go/generated/resourceserver"
)

var (
	environment    *Environment
	unauthedClient *op.Client
	authedClient   *op.AuthenticatedClient
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

	requestBody := as.GrantRequestWithAccessToken{
		AccessToken: accessToken,
	}

	grant, err := authedClient.Grant.Request(
		context.TODO(),
		op.GrantRequestParams{
			URL:         environment.ReceiverOpenPaymentsAuthUrl,
			RequestBody: requestBody,
		},
	)
	if err != nil {
		t.Fatalf("Error with grant request: %v", err)
	}

	printJSON(t, grant)
}

func TestGrantCancel(t *testing.T) {
	grant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error creating incoming payment grant: %v", err)
	}

	err = authedClient.Grant.Cancel(context.TODO(), op.GrantCancelParams{URL: grant.Continue.Uri, AccessToken: grant.Continue.AccessToken.Value})
	if err != nil {
		t.Errorf("Error canceling grant: %v", err)
	}

	// second cancel should error
	err = authedClient.Grant.Cancel(context.TODO(), op.GrantCancelParams{URL: grant.Continue.Uri, AccessToken: grant.Continue.AccessToken.Value})
	if err == nil {
		t.Errorf("Grant cancellation did not error when expected")
	}
}

func TestGrantContinue(t *testing.T) {
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

	url, err := environment.RewriteURLIfNeeded(*newIncomingPayment.Id)
	if err != nil {
		t.Fatalf("Could not rewrite URL from incoming payment: %v", err)
	}

	t.Logf("\nauthedClient.IncomingPayment.Get(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.Get(context.TODO(), op.IncomingPaymentGetParams{
		URL:         url,
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
		BaseURL:       url,
		AccessToken:   grant.AccessToken.Value,
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
		WalletAddressSchema: environment.ResolvedReceiverWalletAddressUrl,
		IncomingAmount: &rs.Amount{
			Value:      "100",
			AssetCode:  environment.ReceiverAssetCode,
			AssetScale: environment.ReceiverAssetScale,
		},
		ExpiresAt: &expiresAt,
		Metadata: &map[string]interface{}{
			"description": "Free Money!",
		},
	}

	incomingPayment, err := authedClient.IncomingPayment.Create(context.TODO(), op.IncomingPaymentCreateParams{
		BaseURL:     url,
		AccessToken: grant.AccessToken.Value,
		Payload:     payload,
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

	url, err := environment.RewriteURLIfNeeded(*incomingPayment.Id)
	if err != nil {
		t.Fatalf("Could not rewrite URL from incoming payment: %v", err)
	}

	// Complete the incoming payment
	t.Logf("\nauthedClient.IncomingPayment.Complete(\"%s\")\n", url)

	completedPayment, err := authedClient.IncomingPayment.Complete(context.TODO(), op.IncomingPaymentCompleteParams{
		URL:         url,
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
	quoteGrantRequestBody := as.GrantRequestWithAccessToken{
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
		WalletAddressSchema: environment.ResolvedSenderWalletAddressUrl,
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
	if retrievedQuote.DebitAmount == (rs.Amount{}) {
		t.Error("Expected retrieved quote to have a debitAmount")
	}
	if retrievedQuote.ReceiveAmount == (rs.Amount{}) {
		t.Error("Expected retrieved quote to have a receiveAmount")
	}
}

func TestCreateAndGetOutgoingPayment(t *testing.T) {
	// Create incoming payment and quote
	incomingPaymentGrant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}
	newIncomingPayment, err := newIncomingPayment(incomingPaymentGrant)
	if err != nil {
		t.Fatalf("Error creating new incoming payment: %v", err)
	}
	newQuote, err := newQuote(newIncomingPayment)
	if err != nil {
		t.Fatalf("Error creating new incoming payment: %v", err)
	}

	// Grant Request
	grant, err := newOutgoingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting outgoing payment grant: %v", err)
	}

	if grant.Interact == nil || grant.Interact.Redirect == "" {
		t.Fatal("Missing interact.redirect URL in grant response")
	}

	// Complete browser interaction
	t.Logf("Opening consent URL: %s", grant.Interact.Redirect)
	err = environment.CompleteConsentFlowWithChromedp(context.TODO(), grant.Interact.Redirect)
	if err != nil {
		t.Fatalf("Error completing browser consent: %v", err)
	}

  // sleep grant.Continue.Wait in seconds
	time.Sleep(time.Duration(*grant.Continue.Wait) * time.Second)

	// Continue Grant
	continuedGrant, err := authedClient.Grant.Continue(context.TODO(), op.GrantContinueParams{
		URL:         grant.Continue.Uri,
		AccessToken: grant.Continue.AccessToken.Value,
	})
	if err != nil {
		t.Fatalf("Error continuing grant: %v", err)
	}

	var paymentPayload rs.CreateOutgoingPaymentRequest
	err = paymentPayload.FromCreateOutgoingPaymentWithQuote(rs.CreateOutgoingPaymentWithQuote{
		WalletAddressSchema: environment.ResolvedSenderWalletAddressUrl,
		QuoteId:             *newQuote.Id,
		Metadata: &map[string]interface{}{
			"purpose": "Integration test",
		},
	})
	if err != nil {
		t.Fatalf("Error creating outgoing payment payload: %v", err)
	}

	newOutgoingPayment, err := authedClient.OutgoingPayment.Create(context.TODO(), op.OutgoingPaymentCreateParams{
		BaseURL:     environment.SenderOpenPaymentsResourceUrl,
		AccessToken: continuedGrant.AccessToken.Value,
		Payload:     paymentPayload,
	})
	if err != nil {
		t.Fatalf("Error creating outgoing payment: %v", err)
	}

	if *newQuote.Id != *newOutgoingPayment.QuoteId {
		t.Errorf("Mismatched quote IDs: got %s, want %s", *newOutgoingPayment.QuoteId, *newQuote.Id)
	}
	if environment.ResolvedSenderWalletAddressUrl != *newOutgoingPayment.WalletAddress {
		t.Errorf("Mismatched wallet addresses: got %s, want %s", *newOutgoingPayment.WalletAddress, environment.ResolvedSenderWalletAddressUrl)
	}

	outgoingPaymentURL, err := environment.RewriteURLIfNeeded(*newOutgoingPayment.Id)
	if err != nil {
		t.Fatalf("Could not rewrite URL from outgoing payment: %v", err)
	}
	t.Logf("\nauthedClient.OutgoingPayment.Get(\"%s\")\n", outgoingPaymentURL)


	retrievedPayment, err := authedClient.OutgoingPayment.Get(context.TODO(), op.OutgoingPaymentGetParams{
		URL:     outgoingPaymentURL,
		AccessToken: continuedGrant.AccessToken.Value,
	})
	if err != nil {
		t.Fatalf("Error getting outgoing payment: %v", err)
	}

	printJSON(t, retrievedPayment)

	if retrievedPayment.Id == nil || *retrievedPayment.Id != *newOutgoingPayment.Id {
		t.Errorf("Expected retrieved outgoing payment ID to be %s, got %s", *newOutgoingPayment.Id, *retrievedPayment.Id)
	}
	if *retrievedPayment.QuoteId != *newOutgoingPayment.QuoteId {
		t.Errorf("Mismatched quote IDs: got %s, want %s", *retrievedPayment.QuoteId, *newOutgoingPayment.QuoteId)
	}
	if *retrievedPayment.WalletAddress != *newOutgoingPayment.WalletAddress {
		t.Errorf("Mismatched wallet addresses: got %s, want %s", *retrievedPayment.WalletAddress, *newOutgoingPayment.WalletAddress)
	}

}


func TestRotateToken(t *testing.T) {
	newIncomingPaymentGrant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}

	originalValue := newIncomingPaymentGrant.AccessToken.Value
	originalManage := newIncomingPaymentGrant.AccessToken.Manage

	rotatedToken, err := authedClient.Token.Rotate(context.TODO(), op.TokenRotateParams{
		URL:         newIncomingPaymentGrant.AccessToken.Manage,
		AccessToken: newIncomingPaymentGrant.AccessToken.Value,
	})
	if err != nil {
		t.Errorf("Error rotating token: %v", err)
	}
	printJSON(t, rotatedToken)

	if rotatedToken.Value == originalValue {
		t.Error("Rotated token value should be different from original")
	}
	if rotatedToken.Manage == originalManage {
		t.Error("Rotated token manage URL should be different from original")
	}

	// can rotate new token
	_, err = authedClient.Token.Rotate(context.TODO(), op.TokenRotateParams{
		URL:         rotatedToken.Manage,
		AccessToken: rotatedToken.Value,
	})
	if err != nil {
		t.Errorf("Error rotating token: %v", err)
	}
	printJSON(t, rotatedToken)

	// cannot rotate revoked token
	_, err = authedClient.Token.Rotate(context.TODO(), op.TokenRotateParams{
		URL:         rotatedToken.Manage,
		AccessToken: rotatedToken.Value,
	})
	if err == nil {
		t.Errorf("Expected error rotating token")
	}
}

func TestRevokeToken(t *testing.T) {
	newIncomingPaymentGrant, err := newIncomingPaymentGrant()
	if err != nil {
		t.Fatalf("Error requesting grant for incoming payment: %v", err)
	}
	err = authedClient.Token.Revoke(context.TODO(), op.TokenRevokeParams{
		URL:         newIncomingPaymentGrant.AccessToken.Manage,
		AccessToken: newIncomingPaymentGrant.AccessToken.Value,
	})
	if err != nil {
		t.Errorf("Error revoking token: %v", err)
	}

	// second revoke should error
	err = authedClient.Token.Revoke(context.TODO(), op.TokenRevokeParams{
		URL:         newIncomingPaymentGrant.AccessToken.Manage,
		AccessToken: newIncomingPaymentGrant.AccessToken.Value,
	})
	if err == nil {
		t.Errorf("Expected error revoking token")
	}
}

// ==============
//
//	Test Helpers
//
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

	requestBody := as.GrantRequestWithAccessToken{
		AccessToken: accessToken,
	}

	grant, err := authedClient.Grant.Request(
		context.TODO(),
		op.GrantRequestParams{
			URL:         environment.ReceiverOpenPaymentsAuthUrl,
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
		WalletAddressSchema: environment.ResolvedReceiverWalletAddressUrl,
		IncomingAmount: &rs.Amount{
			Value:      "100",
			AssetCode:  environment.ReceiverAssetCode,
			AssetScale: environment.ReceiverAssetScale,
		},
		ExpiresAt: &expiresAt,
	}

	incomingPayment, err := authedClient.IncomingPayment.Create(context.TODO(), op.IncomingPaymentCreateParams{
		BaseURL:     url,
		AccessToken: grant.AccessToken.Value,
		Payload:     payload,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating incoming payment: %w", err)
	}

	return &incomingPayment, nil
}

func newQuote(incomingPayment *rs.IncomingPaymentWithMethods) (*rs.Quote, error) {
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
		return nil, fmt.Errorf("error creating AccessItem for quote: %w", err)
	}
	accessToken := struct {
		Access as.Access `json:"access"`
	}{
		Access: []as.AccessItem{accessItem},
	}
	quoteGrantRequestBody := as.GrantRequestWithAccessToken{
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
		return nil, fmt.Errorf("error requesting grant for quote: %w", err)
	}

	// Create the Quote
	createQuotePayload := rs.CreateQuoteJSONBody0{
		WalletAddressSchema: environment.ResolvedSenderWalletAddressUrl,
		Receiver:      *incomingPayment.Id,
		Method:        "ilp",
	}

	quote, err := authedClient.Quote.Create(context.TODO(), op.QuoteCreateParams{
		BaseURL:     environment.SenderOpenPaymentsResourceUrl,
		AccessToken: quoteGrant.AccessToken.Value,
		Payload:     createQuotePayload,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating quote: %w", err)
	}

	return &quote, nil
}

func newOutgoingPaymentGrant() (*op.Grant, error) {
	outgoingAccess := as.AccessOutgoing{
		Type: as.OutgoingPayment,
		Actions: []as.AccessOutgoingActions{
			as.AccessOutgoingActionsCreate,
			as.AccessOutgoingActionsRead,
			as.AccessOutgoingActionsList,
		},
		Identifier: environment.ResolvedSenderWalletAddressUrl,
	}
	accessItem := as.AccessItem{}
	if err := accessItem.FromAccessOutgoing(outgoingAccess); err != nil {
		return nil, fmt.Errorf("error creating AccessItem: %w", err)
	}
	accessToken := struct {
		Access as.Access `json:"access"`
	}{
		Access: []as.AccessItem{accessItem},
	}
	interact := &as.InteractRequest{
		Start: []as.InteractRequestStart{as.InteractRequestStartRedirect},
	}

	grant, err := authedClient.Grant.Request(
		context.TODO(),
		op.GrantRequestParams{
			URL:         environment.SenderOpenPaymentsAuthUrl,
			RequestBody: as.GrantRequestWithAccessToken{
				AccessToken: accessToken,
				Interact:    interact,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error requesting grant: %w", err)
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
