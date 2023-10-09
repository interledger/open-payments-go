package openpayments

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WalletAddress struct{
	httpClient *http.Client
}

// TODO: generate from openapi spec and import instead
type WalletAddressResponse struct {
    ID          string `json:"id"`
    PublicName  string `json:"publicName"`
    AssetCode   string `json:"assetCode"`
    AssetScale  int    `json:"assetScale"`
    AuthServer  string `json:"authServer"`
}

func (wa *WalletAddress) Get(url string) (WalletAddressResponse, error) {
    resp, err := wa.httpClient.Get(url)
    if err != nil {
        return WalletAddressResponse{}, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return WalletAddressResponse{}, fmt.Errorf("failed to get wallet address: %s", resp.Status)
    }

    var walletAddressResponse WalletAddressResponse
    err = json.NewDecoder(resp.Body).Decode(&walletAddressResponse)
    if err != nil {
        return WalletAddressResponse{}, fmt.Errorf("failed to decoding response body: %s", err)
    }

    return walletAddressResponse, nil
}
