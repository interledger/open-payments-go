package openpayments

import (
	// "crypto/ed25519"
	"net/http"

	"github.com/interledger/open-payments-go-sdk/internal/lib"
)

type UnauthenticatedClient struct {
		httpClient *http.Client
    WalletAddress *WalletAddressRoutes
		IncomingPayment *UnauthenticatedIncomingPaymentRoutes
}

func NewUnauthenticatedClient() *UnauthenticatedClient {
		httpClient := http.Client{Transport: &lib.HeaderTransport{Base: http.DefaultTransport}}
    return &UnauthenticatedClient{
				httpClient: &httpClient,
        WalletAddress: &WalletAddressRoutes{httpClient: &httpClient},
				IncomingPayment: &UnauthenticatedIncomingPaymentRoutes{httpClient: &httpClient},
    }
}

// TODO: implement request singing
type AuthenticatedClient struct {
	httpClient      *http.Client
	walletAddressUrl string
	// privateKey      ed25519.PrivateKey
	// keyId           string
	WalletAddress *WalletAddressRoutes
	// Grant           *GrantRoutes
	IncomingPayment *AuthenticatedIncomingPaymentRoutes
}

func NewAuthenticatedClient(walletAddressUrl, privateKey, keyId string) *AuthenticatedClient {
	httpClient := http.Client{Transport: &lib.HeaderTransport{Base: http.DefaultTransport}}
	return &AuthenticatedClient{
		httpClient:      &httpClient,
		walletAddressUrl: walletAddressUrl,
		// privateKey:      edKey,
		// keyId:           keyId,
		WalletAddress: &WalletAddressRoutes{httpClient: &httpClient},
		// Grant:           &GrantRoutes{httpClient: &httpClient},
		IncomingPayment: &AuthenticatedIncomingPaymentRoutes{httpClient: &httpClient},
	}
}