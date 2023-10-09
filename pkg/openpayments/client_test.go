package openpayments

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestWalletAddress_Get_SuccessfulResponse(t *testing.T) {
	expectedWalletAddress := WalletAddressResponse{
		ID:         "https://rafiki.money/alice",
		PublicName: "Alice",
		AssetCode:  "USD",
		AssetScale: 2,
		AuthServer: "https://rafiki.money/auth",
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
	walletAddress, err := client.WalletAddress.Get(server.URL)
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
	walletAddress, err := client.WalletAddress.Get(server.URL)
	if err == nil {
			t.Fatalf("expected an error, got nil")
	}

	if walletAddress != (WalletAddressResponse{}) {
    t.Errorf("expected an empty wallet address, got %+v", walletAddress)
	}

}