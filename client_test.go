package openpayments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	was "github.com/interledger/open-payments-go/generated/walletaddressserver"
)

func TestWalletAddress_Get_SuccessfulResponse(t *testing.T) {
	id := "https://rafiki.money/alice"
	publicName := "Alice"
	assetCode := "USD"
	assetScale := 2
	authServer := "https://rafiki.money/auth"

	expectedWalletAddress := was.WalletAddress{
		Id:             &id,
		PublicName:     &publicName,
		AssetCode:      assetCode,
		AssetScale:     assetScale,
		AuthServer:     &authServer,
		ResourceServer: nil,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse, err := json.Marshal(expectedWalletAddress)
		if err != nil {
			t.Logf("Error marshaling JSON: %v", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponse)
	}))
	defer server.Close()

	client := NewClient()
	walletAddress, err := client.WalletAddress.Get(context.TODO(), WalletAddressGetParams{URL: server.URL})
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !reflect.DeepEqual(walletAddress, expectedWalletAddress) {
		t.Errorf("expected %+v, got %+v", expectedWalletAddress, walletAddress)
	}
}

func TestWalletAddress_Get_FailedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient()
	walletAddress, err := client.WalletAddress.Get(context.TODO(), WalletAddressGetParams{URL: server.URL})
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	if !reflect.DeepEqual(walletAddress, was.WalletAddress{}) {
    t.Errorf("expected an empty wallet address, got %+v", walletAddress)
	}

}
