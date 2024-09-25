package openpayments

import (
	"net/http"

	"github.com/interledger/open-payments-go-sdk/internal/lib"
)

type Client struct {
		httpClient *http.Client
    WalletAddress *WalletAddressRoutes
		IncomingPayment *IncomingPaymentRoutes
		Grant *GrantRoutes
}

func NewClient() *Client {
		httpClient := http.Client{Transport: &lib.HeaderTransport{Base: http.DefaultTransport}}
    return &Client{
				httpClient: &httpClient,
        WalletAddress: &WalletAddressRoutes{httpClient: &httpClient},
				IncomingPayment: &IncomingPaymentRoutes{httpClient: &httpClient},
				Grant: &GrantRoutes{httpClient: &httpClient},
    }
}