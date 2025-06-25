package openpayments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	was "github.com/interledger/open-payments-go/generated/walletaddressserver"
)

type WalletAddressService struct {
	DoUnsigned RequestDoer
}

type WalletAddressGetParams struct {
	URL string // The full URL of the wallet address resource.
}
type WalletAddressGetKeysParams struct {
	URL string // The full URL of the wallet address resource.
}

type WalletAddressGetDIDDocumentParams struct {
	URL string // The full URL of the wallet address resource.
}

func (wa *WalletAddressService) Get(ctx context.Context, params WalletAddressGetParams) (was.WalletAddress, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, params.URL, nil)
	if err != nil {
		return was.WalletAddress{}, err
	}

	resp, err := wa.DoUnsigned(req)
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

func (wa *WalletAddressService) GetKeys(ctx context.Context, params WalletAddressGetKeysParams) (was.JsonWebKeySet, error) {
	url := params.URL + "/jwks.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return was.JsonWebKeySet{}, err
	}

	resp, err := wa.DoUnsigned(req)
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

func (wa *WalletAddressService) GetDIDDocument(ctx context.Context, params WalletAddressGetDIDDocumentParams) (was.DidDocument, error) {
	url := params.URL + "/did.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return was.DidDocument{}, err
	}

	resp, err := wa.DoUnsigned(req)
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
