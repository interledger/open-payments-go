package openpayments_test

import (
	"context"
	"log"
	"net/http"
	"testing"

	openpayments "github.com/interledger/open-payments-go"
	rs "github.com/interledger/open-payments-go/generated/resourceserver"
	"github.com/interledger/open-payments-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestOutgoingPaymentGetGrant_WithSpentAmounts(t *testing.T) {
	mockResponse := openpayments.OutgoingPaymentGrantSpentAmounts{
		SpentReceiveAmount: &rs.Amount{Value: "2500", AssetCode: "USD", AssetScale: 2},
		SpentDebitAmount:   &rs.Amount{Value: "2600", AssetCode: "USD", AssetScale: 2},
	}

	mockServer := testutils.Mock(http.MethodGet, "/outgoing-payment-grant", http.StatusOK, mockResponse)
	defer mockServer.Close()

	client, err := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(mockServer.Client()))
	if err != nil {
		log.Fatalf("Failed to initialize authenticated client: %v", err)
	}

	result, err := client.OutgoingPayment.GetGrantSpentAmounts(context.Background(), openpayments.OutgoingPaymentGrantGetParams{
		BaseURL:     mockServer.URL,
		AccessToken: accessToken,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result.SpentReceiveAmount)
	assert.NotNil(t, result.SpentDebitAmount)
	assert.Equal(t, "2500", result.SpentReceiveAmount.Value)
	assert.Equal(t, "2600", result.SpentDebitAmount.Value)
}

func TestOutgoingPaymentGetGrant_NullSpentAmounts(t *testing.T) {
	mockResponse := map[string]any{
		"spentReceiveAmount": nil,
		"spentDebitAmount":   nil,
	}

	mockServer := testutils.Mock(http.MethodGet, "/outgoing-payment-grant", http.StatusOK, mockResponse)
	defer mockServer.Close()

	client, err := openpayments.NewAuthenticatedClient(walletAddress, pk, keyID, openpayments.WithHTTPClientAuthed(mockServer.Client()))
	if err != nil {
		log.Fatalf("Failed to initialize authenticated client: %v", err)
	}

	result, err := client.OutgoingPayment.GetGrantSpentAmounts(context.Background(), openpayments.OutgoingPaymentGrantGetParams{
		BaseURL:     mockServer.URL,
		AccessToken: accessToken,
	})

	assert.NoError(t, err)
	assert.Nil(t, result.SpentReceiveAmount)
	assert.Nil(t, result.SpentDebitAmount)
}
