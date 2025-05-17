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

func (ip *AuthenticatedIncomingPaymentRoutes) Get(url string, accessToken string) (rs.IncomingPayment, error) {
	incomingPayment, err := lib.FetchAndDecode[rs.IncomingPayment](ip.httpClient, url)
	if err != nil {
		return rs.IncomingPayment{}, fmt.Errorf("failed to get incoming payment: %w", err)
	}
	return incomingPayment, nil
}

type Pagination struct {
	First string
	Last string
	Cursor string
}

type ListArgs struct {
	WalletAddress string
	Pagination Pagination
}

// TODO: wont work because rafiki will fail on open api validation because there is no sig header
func (ip *AuthenticatedIncomingPaymentRoutes) List(url string, accessToken string, args ListArgs) ([]rs.IncomingPayment, error) {
	queryParams := map[string]string{
		"walletAddress": args.WalletAddress,
		"first":         args.Pagination.First,
		"last":          args.Pagination.Last,
		"cursor":        args.Pagination.Cursor,
	}
	fullURL, err := lib.BuildQueryParams(url, queryParams)
	
	if err != nil {
		return nil, fmt.Errorf("failed to build query params: %w", err)
	}

	incomingPayments, err := lib.FetchAndDecode[[]rs.IncomingPayment](ip.httpClient, fullURL)
	if err != nil {
		return []rs.IncomingPayment{}, fmt.Errorf("failed to get incoming payment: %w", err)
	}
	return incomingPayments, nil
}


func getPublic(httpClient *http.Client, url string) (rs.PublicIncomingPayment, error) {
	publicPayment, err := lib.FetchAndDecode[rs.PublicIncomingPayment](httpClient, url)
	if err != nil {
		return rs.PublicIncomingPayment{}, fmt.Errorf("failed to get public incoming payment: %w", err)
	}
	return publicPayment, nil
}
