package main

import (
	"encoding/json"
	"fmt"

	op "github.com/interledger/open-payments-go-sdk/pkg/openpayments"
)

var (
	clientWalletAddress = "http://localhost:4000/accounts/pfry" // clientWalletAddress
	receiverOpenPaymentsAuthHost = "http://localhost:4006"
	unauthedClient = op.NewUnauthenticatedClient()
	authedClient = op.NewAuthenticatedClient("", "", "")
)

func main() {
	// Unauthenticated
	getWalletAddress()
	getWalletAddressKeys()
	getWalletAddressDIDDocument() // Should fail, not implemented in rafiki
	getPublicIncomingPaymentUnauthed("f68ff34f-b052-46a7-b270-71c9dc36e028") // Make payment in rafiki and use id
	
	// Authenticated
	getPublicIncomingPaymentAuthed("f68ff34f-b052-46a7-b270-71c9dc36e028") // Make payment in rafiki and use id
	getIncomingPayment("f68ff34f-b052-46a7-b270-71c9dc36e028") // circumvent signing by hardcoding check in rafiki incoming payment routes
	// grantRequest() // Fails: depends on signing headers to authorize request
}

func getWalletAddress(){
	fmt.Printf("\nclient.WalletAddress.Get(\"%s\")\n", clientWalletAddress)
	walletAddress, err := unauthedClient.WalletAddress.Get(clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address: %v\n", err)
		return
	}

	printJSON(walletAddress)
}

func getWalletAddressKeys(){
	fmt.Printf("\nclient.WalletAddress.GetKeys(\"%s\")\n", clientWalletAddress)
	walletAddressKeys, err := unauthedClient.WalletAddress.GetKeys(clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address keys: %v\n", err)
		return
	}

	printJSON(walletAddressKeys)
}

func getWalletAddressDIDDocument(){
	fmt.Printf("\nclient.WalletAddress.GetDIDDocument(\"%s\"\n)", clientWalletAddress)
	walletAddressDIDDocument, err := unauthedClient.WalletAddress.GetDIDDocument(clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address DID document: %v\n", err)
		return
	}

	printJSON(walletAddressDIDDocument)
}

func getPublicIncomingPaymentUnauthed(incomingPaymentId string) {
	baseUrl := "http://localhost:4000/incoming-payments/"
	url := baseUrl + incomingPaymentId

	fmt.Printf("\nclient.IncomingPayment.GetPublic(\"%s\")\n", url)

	incomingPayment, err := unauthedClient.IncomingPayment.GetPublic(url)
	if err != nil {
		fmt.Printf("Error fetching incoming payment: %v\n", err)
		return
	}

	printJSON(incomingPayment)
}

func getPublicIncomingPaymentAuthed(incomingPaymentId string) {
	baseUrl := "http://localhost:4000/incoming-payments/"
	url := baseUrl + incomingPaymentId

	fmt.Printf("\nclient.IncomingPayment.GetPublic(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.GetPublic(url)
	if err != nil {
		fmt.Printf("Error fetching incoming payment: %v\n", err)
		return
	}

	printJSON(incomingPayment)
}

func getIncomingPayment(incomingPaymentId string) {
	baseUrl := "http://localhost:4000/incoming-payments/"
	url := baseUrl + incomingPaymentId

	fmt.Printf("\nclient.IncomingPayment.Get(\"%s\")\n", url)

	incomingPayment, err := authedClient.IncomingPayment.Get(url)
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
