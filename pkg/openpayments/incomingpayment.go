package openpayments

import (
	"encoding/json"
	"fmt"
	"net/http"

	rs "github.com/interledger/open-payments-go-sdk/pkg/generated/resourceserver"
)

type IncomingPaymentRoutes struct{
	httpClient *http.Client
}

func (ip *IncomingPaymentRoutes) GetPublic(url string) (rs.PublicIncomingPayment, error) {
		resp, err := ip.httpClient.Get(url)
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