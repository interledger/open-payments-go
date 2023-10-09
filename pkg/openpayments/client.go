package openpayments

import (
	"net/http"
)

type Client struct {
		httpClient *http.Client
    WalletAddress *WalletAddress
}

// NewClient creates a new Open Payments client.
func NewClient() *Client {
		httpClient := http.Client{}
    return &Client{
				httpClient: &httpClient,
        WalletAddress: &WalletAddress{httpClient: &httpClient},
    }
}