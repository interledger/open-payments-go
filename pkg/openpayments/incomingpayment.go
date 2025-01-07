package openpayments

import (
	"fmt"
	"net/http"

	"github.com/interledger/open-payments-go-sdk/internal/lib"
	rs "github.com/interledger/open-payments-go-sdk/pkg/generated/resourceserver"
)

type UnauthenticatedIncomingPaymentRoutes struct{
	httpClient *http.Client
}

func (ip *UnauthenticatedIncomingPaymentRoutes) GetPublic(url string) (rs.PublicIncomingPayment, error) {
	return getPublic(ip.httpClient, url)
}

type AuthenticatedIncomingPaymentRoutes struct{
	httpClient *http.Client
}

func (ip *AuthenticatedIncomingPaymentRoutes) GetPublic(url string) (rs.PublicIncomingPayment, error) {
	return getPublic(ip.httpClient, url)
}

func (ip *AuthenticatedIncomingPaymentRoutes) Get(url string) (rs.IncomingPayment, error) {
	incomingPayment, err := lib.FetchAndDecode[rs.IncomingPayment](ip.httpClient, url)
	if err != nil {
		return rs.IncomingPayment{}, fmt.Errorf("failed to get incoming payment: %w", err)
	}
	return incomingPayment, nil
}


func getPublic(httpClient *http.Client, url string) (rs.PublicIncomingPayment, error) {
	publicPayment, err := lib.FetchAndDecode[rs.PublicIncomingPayment](httpClient, url)
	if err != nil {
		return rs.PublicIncomingPayment{}, fmt.Errorf("failed to get public incoming payment: %w", err)
	}
	return publicPayment, nil
}
