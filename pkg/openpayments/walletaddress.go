package openpayments

import (
	"fmt"
	"net/http"

	"github.com/interledger/open-payments-go-sdk/internal/lib"
	was "github.com/interledger/open-payments-go-sdk/pkg/generated/walletaddressserver"
)

type WalletAddressRoutes struct{
	httpClient *http.Client
}

func (wa *WalletAddressRoutes) Get(url string) (was.WalletAddress, error) {
	walletAddress, err := lib.FetchAndDecode[was.WalletAddress](wa.httpClient, url)
	if err != nil {
		return was.WalletAddress{}, fmt.Errorf("failed to get wallet address: %w", err)
	}
	return walletAddress, nil
}

func (wa *WalletAddressRoutes) GetKeys(url string) (was.JsonWebKeySet, error) {
	// Modify the URL directly
	keySet, err := lib.FetchAndDecode[was.JsonWebKeySet](wa.httpClient, url+"/jwks.json")
	if err != nil {
		return was.JsonWebKeySet{}, fmt.Errorf("failed to get json web keys: %w", err)
	}
	return keySet, nil
}

func (wa *WalletAddressRoutes) GetDIDDocument(url string) (was.DidDocument, error) {
	didDocument, err := lib.FetchAndDecode[was.DidDocument](wa.httpClient, url+"/did.json")
	if err != nil {
		return was.DidDocument{}, fmt.Errorf("failed to get DID document: %w", err)
	}
	return didDocument, nil
}
