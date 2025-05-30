package integration

import (
	"net/http"
)

var (
		localPortsToHost = map[string]string{
			"3000": "cloud-nine-wallet-backend",
			"4000": "happy-life-bank-backend",
		}
		localHostsToPort = map[string]string{
			"cloud-nine-wallet-backend": "3000",
			"happy-life-bank-backend": "4000",
		}
)
// TODO: Consider offloading some of this configuration. Not sure if being smarter is worth it.
// - ReceiverOpenPaymentsResourceUrl, auth url, asset code/scale should be discoverable by getting 
//   wallet address url. Do it in TestMain as a sort of configuration step. Perhaps even make it
//   method on Environment? Or part of Environment constructor?
// - util or method on Environment to get the fully resolved urls instead of sorta redundant
//   properties like ReceiverOpenPaymentsResourceUrl

type Environment struct {
	Name                             string
	ClientWalletAddressURL           string
	PrivateKey  					           string
	KeyId                            string
	HttpClient                       *http.Client
	PreSignHook                      func(req *http.Request)
	PostSignHook                     func(req *http.Request)
	ReceiverWalletAddressUrl         string
	ResolvedReceiverWalletAddressUrl string
	ReceiverOpenPaymentsAuthUrl      string
	ReceiverOpenPaymentsResourceUrl  string
	ReceiverAssetScale 							 int
	ReceiverAssetCode                string
}

func NewLocalEnvironment() Environment {
	return Environment{
		Name:                             "local",
		ClientWalletAddressURL:           "https://happy-life-bank-backend/accounts/pfry",
		PrivateKey:                       "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUVxZXptY1BoT0U4Ymt3TitqUXJwcGZSWXpHSWRGVFZXUUdUSEpJS3B6ODgKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo=",
		KeyId:                            "keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5",
		HttpClient:                       &http.Client{
		                                    Transport: &LocalHostHeaderRoundTripper{
                                        rt:           http.DefaultTransport,
                                      },},
		PreSignHook:                      LocalPreSignHook,
		PostSignHook:                     LocalPostSignHook,
		ResolvedReceiverWalletAddressUrl: "https://happy-life-bank-backend/accounts/pfry",
		ReceiverWalletAddressUrl:         "http://localhost:4000/accounts/pfry",
    ReceiverOpenPaymentsAuthUrl:      "http://localhost:4006",
		ReceiverOpenPaymentsResourceUrl:  "http://localhost:4000",
		ReceiverAssetCode:                 "USD",
		ReceiverAssetScale: 							 2,
	}
}

// TODO: NewTestnetEnvironment
// func NewTestnetEnvironment() Environment {
// 	return Environment{
// 		Name:                            "testnet",
// 		ClientWalletAddressURL:          "", // read from ENV, like "https://interledger-test.dev/blair"
// 		PrivateKey:                      "", // read from ENV
// 		KeyId:                           "", // read from ENV
// 		HttpClient:                      &http.Client{},
// 		PreSignHook:                     nil,
// 		PostSignHook:                    nil,
// 		ReceiverWalletAddressUrl:        "",
//    ReceiverOpenPaymentsAuthUrl:     "https://auth.interledger-test.dev/",
// 		ReceiverOpenPaymentsResourceUrl: "",
// 	}
// }

// HostHeaderRoundTripper modifies Host header to match remote services while using localhost.
// RoundTripper will modify all requests after DoSigned/DoUnsigned
type LocalHostHeaderRoundTripper struct {
	rt http.RoundTripper
}

func (h *LocalHostHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Hostname() == "localhost" {
		port := req.URL.Port()
		if host, ok := localPortsToHost[port]; ok {
			req.Host = host
		}
	}
	
	return h.rt.RoundTrip(req)
}


func LocalPreSignHook(req *http.Request) {
	// Use resolved URL for signing to match how backend forms url. 
	// Mirrors sanitization done by bruno before signing requests.
	// Note: this should not map auth urls (for exmaple localhost:4006)
	if req.URL.Hostname() == "localhost" {
		port := req.URL.Port()
		if host, ok := localPortsToHost[port]; ok {
			req.URL.Host = host
		}
	}
}

func LocalPostSignHook(req *http.Request) {
	// Restore URL - only needs to be resolved version for signing.
	host := req.URL.Hostname()
	if port, ok := localHostsToPort[host]; ok {
		req.URL.Host = "localhost:" + port
	}
}