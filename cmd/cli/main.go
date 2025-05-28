package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	as "github.com/interledger/open-payments-go-sdk/pkg/generated/authserver"
	op "github.com/interledger/open-payments-go-sdk/pkg/openpayments"
)

var (
	clientWalletAddress = "http://localhost:4000/accounts/pfry"
	receiverOpenPaymentsAuthHost = "http://localhost:4006"
	client = op.NewClient(op.WithHTTPClientUnauthed(&http.Client{
		Transport: &HostHeaderRoundTripper{
			rt:           http.DefaultTransport,
		},
	}),)
)

var authedClient = op.NewAuthenticatedClient(
	"http://localhost:4000/accounts/pfry", 
	// for testnet (get key from interledger-test.dev)
	"LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUVxZXptY1BoT0U4Ymt3TitqUXJwcGZSWXpHSWRGVFZXUUdUSEpJS3B6ODgKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo=",
	"keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5",
	op.WithHTTPClientAuthed(&http.Client{
		Transport: &HostHeaderRoundTripper{
			rt:           http.DefaultTransport,
		},
	}),
)

// HostHeaderRoundTripper is a RoundTripper used in this test to coerce 
// the localhost host header to cloud-nine-wallet-backend/happy-life-bank-backend.
// This prevents the wallet address middleware from incorrectly forming
// the wallet address using localhost and causing failure(s?) such as 
// GetWalletAddressKeys returning empty keys for a new wallet address instead 
// of the actual keys from the existing address. This round tripper mirrors a 
// bruno pre-request script.
type HostHeaderRoundTripper struct {
	rt       http.RoundTripper
}

// RoundTrip modifies the Host header if the request is to localhost:3000 or 4000.
func (h *HostHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	portsToHost := map[string]string{
    "3000": "cloud-nine-wallet-backend",
    "4000": "happy-life-bank-backend",
	}

	if req.URL.Hostname() == "localhost" {
		port := req.URL.Port()
		if host, ok := portsToHost[port]; ok { 
			req.Host = host
		}
	}
	return h.rt.RoundTrip(req)
}

func main() {
	// Unauthenticated
	getWalletAddress()
	getWalletAddressKeys()
	getWalletAddressDIDDocument() // Should fail, not implemented in rafiki
	getPublicIncomingPaymentUnauthed("c021ed69-45fe-4bf3-9e2a-27c5bb6b0131") // Make payment in rafiki and use id

	// Authenticated
	getPublicIncomingPaymentAuthed("c021ed69-45fe-4bf3-9e2a-27c5bb6b0131")   // Make payment in rafiki and use id
	getIncomingPayment("c021ed69-45fe-4bf3-9e2a-27c5bb6b0131")               // Make payment in rafiki and use id
	getIncomingPaymentList()
	grantRequest()
}

func getWalletAddress(){
	fmt.Printf("\nclient.WalletAddress.Get(\"%s\")\n", clientWalletAddress)
	walletAddress, err := client.WalletAddress.Get(context.TODO(), clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address: %v\n", err)
		return
	}

	printJSON(walletAddress)
}

func getWalletAddressKeys(){
	fmt.Printf("\nclient.WalletAddress.GetKeys(\"%s\")\n", clientWalletAddress)
	walletAddressKeys, err := client.WalletAddress.GetKeys(context.TODO(), clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address keys: %v\n", err)
		return
	}

	printJSON(walletAddressKeys)
}

func getWalletAddressDIDDocument(){
	fmt.Printf("\nclient.WalletAddress.GetDIDDocument(\"%s\"\n)", clientWalletAddress)
	walletAddressDIDDocument, err := client.WalletAddress.GetDIDDocument(context.TODO(), clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address DID document: %v\n", err)
		return
	}

	printJSON(walletAddressDIDDocument)
}

func getPublicIncomingPaymentUnauthed(incomingPaymentId string) {
	baseUrl := "http://localhost:4000/incoming-payments/"
	url := baseUrl + incomingPaymentId

	fmt.Printf("\nclient.IncomingPayment.GetPublic(\"%s\"\n)", url)

	incomingPayment, err := client.IncomingPayment.GetPublic(context.TODO(), url)
	if err != nil {
		fmt.Printf("Error fetching incoming payment: %v\n", err)
		return
	}

	printJSON(incomingPayment)
}

func getPublicIncomingPaymentAuthed(incomingPaymentId string) {
	baseUrl := "http://localhost:4000/incoming-payments/"
	url := baseUrl + incomingPaymentId

	fmt.Printf("\nclient.IncomingPayment.GetPublic(\"%s\"\n)", url)

	incomingPayment, err := authedClient.IncomingPayment.GetPublic(context.TODO(), url)
	if err != nil {
		fmt.Printf("Error fetching incoming payment: %v\n", err)
		return
	}

	printJSON(incomingPayment)
}

// grant request quote
// func grantRequest(){
// 	// access token
// 	quoteAccess := as.AccessQuote{
// 		Type:    as.Quote,
// 		Actions: []as.AccessQuoteActions{as.Create, as.Read},
// 	}
// 	accessItem := as.AccessItem{}
// 	err := accessItem.FromAccessQuote(quoteAccess)
// 	if err != nil {
// 		fmt.Println("Error creating AccessItem:", err)
// 		return
// 	}
// 	accessToken := struct {
// 		Access as.Access `json:"access"` // TODO: remove this json bit?
// 	}{
// 		Access: []as.AccessItem{accessItem},
// 	}

// 	// interact
// 	// interact := as.InteractRequest{
// 	// 	Start: []as.InteractRequestStart{as.InteractRequestStartRedirect},
// 	// 	Finish: &struct{
// 	// 		Method as.InteractRequestFinishMethod "json:\"method\"";
// 	// 		Nonce string "json:\"nonce\"";
// 	// 		Uri string "json:\"uri\"";
// 	// 	}{
// 	// 		Method: as.InteractRequestFinishMethodRedirect,
// 	// 		Nonce: "456",
// 	// 		Uri: "http://localhost:3030/mock-idp/fake-client",
// 	// 	},
// 	// }

// 	requestBody := as.PostRequestJSONBody{
// 		AccessToken: accessToken,
// 		// Client should be, clientWalletAddress but adjusted to use domain name
// 		// Client:      clientWalletAddress, //"https://happy-life-bank-backend/accounts/pfry"
// 		Client:      "https://happy-life-bank-backend/accounts/pfry",
// 		// Interact: &interact,
// 	}
// 	fmt.Printf("\nclient.Grant.Request(\"%s\", \"%+v\"\n)", clientWalletAddress, requestBody)
// 	grantRequest, err := authedClient.Grant.Request(receiverOpenPaymentsAuthHost, requestBody)
	
// 	if err != nil {
// 		fmt.Printf("Error with grant request: %v\n", err)
// 		return
// 	}
// 	printJSON(grantRequest)
// }

// grant request incoming payment
func grantRequest() {
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
	err := accessItem.FromAccessIncoming(incomingAccess)
	if err != nil {
			fmt.Println("Error creating AccessItem:", err)
			return
	}
	accessToken := struct {
			Access as.Access `json:"access"`
	}{
			Access: []as.AccessItem{accessItem},
	}
	requestBody := as.PostRequestJSONBody{
			AccessToken: accessToken,
			// for local
			Client:      "https://happy-life-bank-backend/accounts/pfry",
			// for testnet:
			// Client:      		"https://interledger-test.dev/blair", // or similar
	}
	
	// for testnet
	// errors: key is not valid base64?
	// grantRequest, err := authedClient.Grant.Request("https://auth.interledger-test.dev/", requestBody)
	
	grant, err := authedClient.Grant.Request(context.TODO(), receiverOpenPaymentsAuthHost, requestBody)

	if err != nil {
			fmt.Printf("Error with grant request: %v\n", err)
			return
	}

	fmt.Println("Completed grant request with DoSigned")

	printJSON(grant)
}

func getIncomingPayment(incomingPaymentId string) {
	baseUrl := "http://localhost:4000/incoming-payments/"
	url := baseUrl + incomingPaymentId

	fmt.Printf("\nclient.IncomingPayment.Get(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.Get(context.TODO(), url, "")
	if err != nil {
		fmt.Printf("Error fetching incoming payment: %v\n", err)
		return
	}

	printJSON(incomingPayment)
}

func getIncomingPaymentList() {
	url := "http://localhost:4000/incoming-payments/"
	// url := baseUrl + incomingPaymentId

	fmt.Printf("\nclient.IncomingPayment.List(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.List(context.TODO(), url, "", op.ListArgs{
		WalletAddress: "some-wallet-address",
		Pagination: op.Pagination{
			First:  "10",
			Cursor: "abc123",
		},
	})
	if err != nil {
		fmt.Printf("Error fetching incoming payment: %v\n", err)
		return
	}

	printJSON(incomingPayment)
}

// func grantRequest(){
// 	// access token
// 	quoteAccess := as.AccessQuote{
// 		Type:    as.Quote,
// 		Actions: []as.AccessQuoteActions{as.Create, as.Read},
// 	}
// 	accessItem := as.AccessItem{}
// 	err := accessItem.FromAccessQuote(quoteAccess)
// 	if err != nil {
// 		fmt.Println("Error creating AccessItem:", err)
// 		return
// 	}
// 	accessToken := struct {
// 		Access as.Access `json:"access"` // TODO: remove this json bit?
// 	}{
// 		Access: []as.AccessItem{accessItem},
// 	}

// 	// interact
// 	// interact := as.InteractRequest{
// 	// 	Start: []as.InteractRequestStart{as.InteractRequestStartRedirect},
// 	// 	Finish: &struct{
// 	// 		Method as.InteractRequestFinishMethod "json:\"method\"";
// 	// 		Nonce string "json:\"nonce\"";
// 	// 		Uri string "json:\"uri\"";
// 	// 	}{
// 	// 		Method: as.InteractRequestFinishMethodRedirect,
// 	// 		Nonce: "456",
// 	// 		Uri: "http://localhost:3030/mock-idp/fake-client",
// 	// 	},
// 	// }

// 	requestBody := as.PostRequestJSONBody{
// 		AccessToken: accessToken,
// 		Client:      clientWalletAddress,
// 		// Interact: &interact,
// 	}
// 	fmt.Printf("\nclient.Grant.Request(\"%s\", \"%+v\"\n)", clientWalletAddress, requestBody)
// 	grantRequest, err := client.Grant.Request(receiverOpenPaymentsAuthHost, requestBody)
// 	if err != nil {
// 		fmt.Printf("Error with grant request: %v\n", err)
// 		return
// 	}
// 	printJSON(grantRequest)

// }

func printJSON(v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}
