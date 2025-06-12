package integration

import (
	"net/http"
	"net/url"
)

// TODO: Consider offloading some of this configuration to using URLs (auth server url, resource
// server url, etc.) returned by wallet address get, etc. in test. May be smarter but not as easy
// to manage as this static configuration .. ?

type Environment struct {
	Name                             string
	ClientWalletAddressURL           string
	PrivateKey  					           string
	KeyId                            string
	HttpClient                       *http.Client
	ReceiverWalletAddressUrl         string
	ResolvedReceiverWalletAddressUrl string
	ReceiverOpenPaymentsAuthUrl      string
	ReceiverOpenPaymentsResourceUrl  string
	ReceiverAssetScale 							 int
	ReceiverAssetCode                string
	SenderOpenPaymentsAuthUrl        string
	SenderWalletAddressUrl           string
	ResolvedSenderWalletAddressUrl   string
	SenderOpenPaymentsResourceUrl    string
	PreSignHook                      func(req *http.Request)
	PostSignHook                     func(req *http.Request)
	RewriteURL func(string) (string, error) // optional, used only in local
}

func NewLocalEnvironment() *Environment {
	localPortsToHost := map[string]string{
		"3000": "cloud-nine-wallet-backend",
		"4000": "happy-life-bank-backend",
	}
	localHostsToPort := map[string]string{
		"cloud-nine-wallet-backend": "3000",
		"happy-life-bank-backend":   "4000",
	}

	env := Environment{
		Name:                             "local",
		ClientWalletAddressURL:           "https://happy-life-bank-backend/accounts/pfry",
		PrivateKey:                       "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUVxZXptY1BoT0U4Ymt3TitqUXJwcGZSWXpHSWRGVFZXUUdUSEpJS3B6ODgKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo=",
		KeyId:                            "keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5",
		HttpClient:                       &http.Client{
																				Transport: MakeLocalHostHeaderRoundTripper(localPortsToHost),
																			},
		ResolvedReceiverWalletAddressUrl: "https://happy-life-bank-backend/accounts/pfry",
		ReceiverWalletAddressUrl:         "http://localhost:4000/accounts/pfry",
		ReceiverOpenPaymentsAuthUrl:      "http://localhost:4006",
		ReceiverOpenPaymentsResourceUrl:  "http://localhost:4000",
		ReceiverAssetCode:                "USD",
		ReceiverAssetScale:               2,
		SenderOpenPaymentsAuthUrl:        "http://localhost:3006",
		SenderOpenPaymentsResourceUrl:    "http://localhost:3000",
		SenderWalletAddressUrl:           "http://localhost:3000/accounts/gfranklin",
		ResolvedSenderWalletAddressUrl:   "https://cloud-nine-wallet-backend/accounts/gfranklin",
		PreSignHook: 											MakeLocalPreSignHook(localPortsToHost),
		PostSignHook: 										MakeLocalPostSignHook(localHostsToPort),
		RewriteURL: 											MakeLocalURLRewriter(localHostsToPort),
	}

	return &env
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
	rt            http.RoundTripper
	portsToHost   map[string]string
}

func (h *LocalHostHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Hostname() == "localhost" {
		port := req.URL.Port()
		if host, ok := h.portsToHost[port]; ok {
			req.Host = host
		}
	}
	return h.rt.RoundTrip(req)
}

func MakeLocalHostHeaderRoundTripper(portsToHost map[string]string) http.RoundTripper {
	return &LocalHostHeaderRoundTripper{
		rt:          http.DefaultTransport,
		portsToHost: portsToHost,
	}
}

func MakeLocalPreSignHook(portsToHost map[string]string) func(req *http.Request) {
	return func(req *http.Request) {
		if req.URL.Hostname() == "localhost" {
			if port := req.URL.Port(); port != "" {
				if host, ok := portsToHost[port]; ok {
					req.URL.Host = host
				}
			}
		}
	}
}

func MakeLocalPostSignHook(hostsToPort map[string]string) func(req *http.Request) {
	return func(req *http.Request) {
		if port, ok := hostsToPort[req.URL.Hostname()]; ok {
			req.URL.Host = "localhost:" + port
		}
	}
}

// ResolveLocalURL takes a URL and replaces known backend hostnames with localhost and their corresponding ports.
// for example: 
// 		"http://happy-life-bank-backend/incoming-payments/f6eabfa0-3a94-4ae6-a635-bf43f9af3aee"
// goes to 
//		"http://localhost:4000/incoming-payments/f6eabfa0-3a94-4ae6-a635-bf43f9af3aee"
func MakeLocalURLRewriter(hostsToPort map[string]string) func(string) (string, error) {
	return func(raw string) (string, error) {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", err
		}

		if port, ok := hostsToPort[parsed.Hostname()]; ok {
			parsed.Host = "localhost:" + port
		}

		return parsed.String(), nil
	}
}

func (env *Environment) RewriteURLIfNeeded(raw string) (string, error) {
	if env.RewriteURL != nil {
		return env.RewriteURL(raw)
	}
	return raw, nil
}
