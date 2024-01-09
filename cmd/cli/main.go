package main

import (
	"encoding/json"
	"fmt"

	op "github.com/interledger/open-payments-go-sdk/pkg/openpayments"
)

var (
	walletAddressUrl = "http://localhost:4000/accounts/pfry"
	client = op.NewClient()
)

func main() {
	getWalletAddress()
	getWalletAddressKeys()
	getWalletAddressDIDDocument() // Should fail, not implement in rafiki

}

func getWalletAddress(){
	fmt.Printf("\nclient.WalletAddress.Get(\"%s\")\n", walletAddressUrl)
	walletAddress, err := client.WalletAddress.Get(walletAddressUrl)

	if err != nil {
		fmt.Printf("Error fetching wallet address: %v\n", err)
		return
	}

	printJSON(walletAddress)
}

func getWalletAddressKeys(){
	fmt.Printf("\nclient.WalletAddress.GetKeys(\"%s\")\n", walletAddressUrl)
	walletAddressKeys, err := client.WalletAddress.GetKeys(walletAddressUrl)

	if err != nil {
		fmt.Printf("Error fetching wallet address keys: %v\n", err)
		return
	}

	printJSON(walletAddressKeys)
}

func getWalletAddressDIDDocument(){
	fmt.Printf("\nclient.WalletAddress.GetDIDDocument(\"%s\"\n)", walletAddressUrl)
	walletAddressDIDDocument, err := client.WalletAddress.GetDIDDocument(walletAddressUrl)

	if err != nil {
		fmt.Printf("Error fetching wallet address DID document: %v\n", err)
		return
	}

	printJSON(walletAddressDIDDocument)
}

func printJSON(v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}
	fmt.Println(string(jsonData))
}
