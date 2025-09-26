package integration

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/joho/godotenv"
)

// TODO: Consider offloading some of this configuration to using URLs (auth server url, resource
// server url, etc.) returned by wallet address get, etc. in test. May be smarter but not as easy
// to manage as this static configuration .. ?

type Environment struct {
	Name                             string
	ClientWalletAddressURL           string
	PrivateKey                       string
	KeyId                            string
	HttpClient                       *http.Client
	ReceiverWalletAddressUrl         string
	ResolvedReceiverWalletAddressUrl string
	ReceiverOpenPaymentsAuthUrl      string
	ReceiverOpenPaymentsResourceUrl  string
	ReceiverAssetScale               int
	ReceiverAssetCode                string
	SenderOpenPaymentsAuthUrl        string
	SenderWalletAddressUrl           string
	ResolvedSenderWalletAddressUrl   string
	SenderOpenPaymentsResourceUrl    string
	PreSignHook                      func(req *http.Request)
	PostSignHook                     func(req *http.Request)
	RewriteURL                       func(string) (string, error) // optional, used only in local
	Consent                          func(ctx context.Context, url string) error
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
		Name:                   "local",
		ClientWalletAddressURL: "https://happy-life-bank-backend/accounts/pfry",
		PrivateKey:             "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUVxZXptY1BoT0U4Ymt3TitqUXJwcGZSWXpHSWRGVFZXUUdUSEpJS3B6ODgKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo=",
		KeyId:                  "keyid-97a3a431-8ee1-48fc-ac85-70e2f5eba8e5",
		HttpClient: &http.Client{
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
		PreSignHook:                      MakeLocalPreSignHook(localPortsToHost),
		PostSignHook:                     MakeLocalPostSignHook(localHostsToPort),
		RewriteURL:                       MakeLocalURLRewriter(localHostsToPort),
		Consent:                          MakeLocalConsent(),
	}

	return &env
}

func NewTestnetEnvironment() *Environment {
	// load env vars and error if required are missing
	file := ".env.testnet"
	if err := godotenv.Load(file); err != nil {
		log.Fatalf("error loading %s: %v", file, err)
	}
	requiredVars := []string{
	"KEY_ID",
	"PRIVATE_KEY_BASE64",
	"SENDING_WALLET_ADDRESS",
	"SENDING_WALLET_ADDRESS_EMAIL",
	"SENDING_WALLET_ADDRESS_PASSWORD",
	}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("required environment variable %s not set in %s", v, file)
		}
	}

	sendingWalletAddress := os.Getenv("SENDING_WALLET_ADDRESS")

	env := Environment{
		Name:                   					"testnet",
		ClientWalletAddressURL: 					sendingWalletAddress,
		PrivateKey:             					os.Getenv("PRIVATE_KEY_BASE64"),
		KeyId:                  					os.Getenv("KEY_ID"),
		HttpClient:                      	&http.Client{},
		PreSignHook:                     	nil,
		PostSignHook:                    	nil,
		ResolvedReceiverWalletAddressUrl: "https://ilp.interledger-test.dev/blair",
		ReceiverWalletAddressUrl:         "https://ilp.interledger-test.dev/blair",
		ReceiverOpenPaymentsAuthUrl:      "https://auth.interledger-test.dev",
		ReceiverOpenPaymentsResourceUrl:  "https://ilp.interledger-test.dev",
		ReceiverAssetCode:                "USD",
		ReceiverAssetScale:               2,
		SenderOpenPaymentsAuthUrl:        "https://auth.interledger-test.dev",
		SenderOpenPaymentsResourceUrl:    "https://ilp.interledger-test.dev",
		SenderWalletAddressUrl:           sendingWalletAddress,
		ResolvedSenderWalletAddressUrl:   sendingWalletAddress,
		Consent:                          MakeTestnetConsent(),
	}

	return &env
}

// HostHeaderRoundTripper modifies Host header to match remote services while using localhost.
// RoundTripper will modify all requests after DoSigned/DoUnsigned
type LocalHostHeaderRoundTripper struct {
	rt          http.RoundTripper
	portsToHost map[string]string
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
//
//	"http://happy-life-bank-backend/incoming-payments/f6eabfa0-3a94-4ae6-a635-bf43f9af3aee"
//
// goes to
//
//	"http://localhost:4000/incoming-payments/f6eabfa0-3a94-4ae6-a635-bf43f9af3aee"
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

func makeConsent(url string, actions func(page *rod.Page, snapshot *Snapshotter) error) error {
	u := launcher.New().
		Headless(true).
		NoSandbox(true).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage().Timeout(30 * time.Second)
	snapshotter := NewSnapshotter(page)

	err := rod.Try(func() {
		page.MustNavigate(url).MustWaitLoad()
		snapshotter.Snapshot()
		if err := actions(page, snapshotter); err != nil {
			panic(err)
		}
	})

	if err != nil {
		fmt.Println("Error interacting with consent UI:", err)
		fmt.Println("Writing last snapshots to disk...")
		snapshotter.FlushOnError()
		return fmt.Errorf("error interacting with consent UI: %w", err)
	}

	return nil
}

func MakeLocalConsent() func(ctx context.Context, url string) error {
	return func(ctx context.Context, url string) error {
		return makeConsent(url, func(page *rod.Page, snapshotter *Snapshotter) error {
			fmt.Println("Waiting for 'allow' button...")
			page.MustElement(`button[aria-label="allow"]`).MustWaitVisible().MustClick()
			fmt.Println("Clicked allow.")
			snapshotter.Snapshot()

			fmt.Println("Waiting for 'close' button...")
			page.MustElement(`button[aria-label="close"]`).MustWaitVisible().MustClick()
			fmt.Println("Clicked close.")
			snapshotter.Snapshot()

			page.MustElementR("*", "Accepted")
			fmt.Println("Consent accepted.")
			snapshotter.Snapshot()
			return nil
		})
	}
}

func MakeTestnetConsent() func(ctx context.Context, url string) error {
	return func(ctx context.Context, url string) error {
		return makeConsent(url, func(page *rod.Page, snapshotter *Snapshotter) error {
			fmt.Println("Waiting for login screen...")
			page.MustElement(`input[type="email"]`).MustWaitVisible().MustInput(os.Getenv("SENDING_WALLET_ADDRESS_EMAIL"))
			page.MustElement(`input[type="password"]`).MustWaitVisible().MustInput(os.Getenv("SENDING_WALLET_ADDRESS_PASSWORD"))
			page.MustElement(`button[aria-label="login"]`).MustWaitVisible().MustClick()
			fmt.Println("Input credentials and clicked login.")
			snapshotter.Snapshot()

			fmt.Println("Waiting for 'accept' button...")
			page.MustElement(`button[aria-label="accept"]`).MustWaitVisible().MustClick()
			fmt.Println("Clicked accept.")
			snapshotter.Snapshot()

			page.MustElementR("*", "Accepted")
			fmt.Println("Consent accepted.")
			snapshotter.Snapshot()
			return nil
		})
	}
}

type Snapshotter struct {
	Page           *rod.Page
	LastScreenshot []byte
	LastHTML       string
}

func NewSnapshotter(page *rod.Page) *Snapshotter {
	return &Snapshotter{Page: page}
}

func (s *Snapshotter) Snapshot() {
	if img, err := s.Page.Screenshot(true, nil); err == nil {
		s.LastScreenshot = img
	} else {
		fmt.Println("Warning: failed to capture screenshot:", err)
	}

	if html, err := s.Page.HTML(); err == nil {
		s.LastHTML = html
	} else {
		fmt.Println("Warning: failed to capture html:", err)
	}
}

func (s *Snapshotter) FlushOnError() {
	const artifactDir = "artifacts"
	os.MkdirAll(artifactDir, 0755)

	if s.LastScreenshot != nil {
		if err := os.WriteFile("artifacts/error_screenshot.png", s.LastScreenshot, 0644); err != nil {
			fmt.Println("Failed to save error screenshot:", err)
		}
	} else {
		fmt.Println("No screenshot captured before error.")
	}

	if s.LastHTML != "" {
		if err := os.WriteFile("artifacts/error_page.html", []byte(s.LastHTML), 0644); err != nil {
			fmt.Println("Failed to save error page html:", err)
		}
	} else {
		fmt.Println("No html captured before error.")
	}
}