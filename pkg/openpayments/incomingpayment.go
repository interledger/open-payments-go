package openpayments

import (
	"encoding/json"
	"fmt"
	"net/http"

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

// func (ip *UnauthenticatedIncomingPaymentRoutes) Get(url string) (rs.IncomingPayment, error) {
// 	// TODO: implement this
// }

func getPublic(httpClient *http.Client, url string) (rs.PublicIncomingPayment, error) {
resp, err := httpClient.Get(url)
		if err != nil {
				return rs.PublicIncomingPayment{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
				return rs.PublicIncomingPayment{}, fmt.Errorf("failed to get incoming payment: %s", resp.Status)
		}

		var incomingPaymentResponse rs.PublicIncomingPayment
		err = json.NewDecoder(resp.Body).Decode(&incomingPaymentResponse)
		if err != nil {
				return rs.PublicIncomingPayment{}, fmt.Errorf("failed to decoding response body: %s", err)
		}

		return incomingPaymentResponse, nil
}