package main

import (
	"encoding/json"
	"fmt"

	as "github.com/interledger/open-payments-go-sdk/pkg/generated/authserver"
	op "github.com/interledger/open-payments-go-sdk/pkg/openpayments"
)

var (
	clientWalletAddress = "http://localhost:4000/accounts/pfry" // clientWalletAddress
	receiverOpenPaymentsAuthHost = "http://localhost:4006"
	client = op.NewClient()
)

func main() {
	getWalletAddress()
	getWalletAddressKeys()
	getWalletAddressDIDDocument() // Should fail, not implement in rafiki
	grantRequest() // Fails: depends on signing headers to authorize request
}

func getWalletAddress(){
	fmt.Printf("\nclient.WalletAddress.Get(\"%s\")\n", clientWalletAddress)
	walletAddress, err := client.WalletAddress.Get(clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address: %v\n", err)
		return
	}

	printJSON(walletAddress)
}

func getWalletAddressKeys(){
	fmt.Printf("\nclient.WalletAddress.GetKeys(\"%s\")\n", clientWalletAddress)
	walletAddressKeys, err := client.WalletAddress.GetKeys(clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address keys: %v\n", err)
		return
	}

	printJSON(walletAddressKeys)
}

func getWalletAddressDIDDocument(){
	fmt.Printf("\nclient.WalletAddress.GetDIDDocument(\"%s\"\n)", clientWalletAddress)
	walletAddressDIDDocument, err := client.WalletAddress.GetDIDDocument(clientWalletAddress)

	if err != nil {
		fmt.Printf("Error fetching wallet address DID document: %v\n", err)
		return
	}

	printJSON(walletAddressDIDDocument)
}

func grantRequest(){
	// access token
	quoteAccess := as.AccessQuote{
		Type:    as.Quote,
		Actions: []as.AccessQuoteActions{as.Create, as.Read},
	}
	accessItem := as.AccessItem{}
	err := accessItem.FromAccessQuote(quoteAccess)
	if err != nil {
		fmt.Println("Error creating AccessItem:", err)
		return
	}
	accessToken := struct {
		Access as.Access `json:"access"` // TODO: remove this json bit?
	}{
		Access: []as.AccessItem{accessItem},
	}

	// interact
	// interact := as.InteractRequest{
	// 	Start: []as.InteractRequestStart{as.InteractRequestStartRedirect},
	// 	Finish: &struct{
	// 		Method as.InteractRequestFinishMethod "json:\"method\"";
	// 		Nonce string "json:\"nonce\"";
	// 		Uri string "json:\"uri\"";
	// 	}{
	// 		Method: as.InteractRequestFinishMethodRedirect,
	// 		Nonce: "456",
	// 		Uri: "http://localhost:3030/mock-idp/fake-client",
	// 	},
	// }

	requestBody := as.PostRequestJSONBody{
		AccessToken: accessToken,
		Client:      clientWalletAddress,
		// Interact: &interact,
	}
	fmt.Printf("\nclient.Grant.Request(\"%s\", \"%+v\"\n)", clientWalletAddress, requestBody)
	grantRequest, err := client.Grant.Request(receiverOpenPaymentsAuthHost, requestBody)
	if err != nil {
		fmt.Printf("Error with grant request: %v\n", err)
		return
	}
	printJSON(grantRequest)

}

func printJSON(v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}
