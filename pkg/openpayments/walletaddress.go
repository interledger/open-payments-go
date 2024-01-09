package openpayments

import (
	"encoding/json"
	"fmt"
	"net/http"

	was "github.com/interledger/open-payments-go-sdk/pkg/generated/walletaddressserver"
)

type WalletAddressControllers struct{
	httpClient *http.Client
}

func (wa *WalletAddressControllers) Get(url string) (was.WalletAddress, error) {
    resp, err := wa.httpClient.Get(url)
    if err != nil {
        return was.WalletAddress{}, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return was.WalletAddress{}, fmt.Errorf("failed to get wallet address: %s", resp.Status)
    }

    var walletAddressResponse was.WalletAddress
    err = json.NewDecoder(resp.Body).Decode(&walletAddressResponse)
    if err != nil {
        return was.WalletAddress{}, fmt.Errorf("failed to decoding response body: %s", err)
    }

    return walletAddressResponse, nil
}

func (wa *WalletAddressControllers) GetKeys(url string) (was.JsonWebKeySet, error) {
    resp, err := wa.httpClient.Get(url + "/jwks.json")
    if err != nil {
        return was.JsonWebKeySet{}, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return was.JsonWebKeySet{}, fmt.Errorf("failed to get json web keys: %s", resp.Status)
    }

    var keyResponse was.JsonWebKeySet
    err = json.NewDecoder(resp.Body).Decode(&keyResponse)
    if err != nil {
        return was.JsonWebKeySet{}, fmt.Errorf("failed to decoding response body: %s", err)
    }

    return keyResponse, nil
}

func (wa *WalletAddressControllers) GetDIDDocument(url string) (was.DidDocument, error) {
    resp, err := wa.httpClient.Get(url + "/did.json")
    if err != nil {
        return was.DidDocument{}, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return was.DidDocument{}, fmt.Errorf("failed to get DID document: %s", resp.Status)
    }

    var DIDDocumentResponse was.DidDocument
    err = json.NewDecoder(resp.Body).Decode(&DIDDocumentResponse)
    if err != nil {
        return was.DidDocument{}, fmt.Errorf("failed to decoding response body: %s", err)
    }

    return DIDDocumentResponse, nil
}