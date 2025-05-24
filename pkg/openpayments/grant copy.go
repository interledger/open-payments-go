package openpayments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	as "github.com/interledger/open-payments-go-sdk/pkg/generated/authserver"
)

// TODO: new Grant routes/service. Use when original grant.go is removed.

type GrantService struct {
	DoSigned   RequestDoer
}

type Grant2 struct {
	AccessToken as.AccessToken
	Continue    as.Continue
}

// TODO: fix return:
// For bruno example this has a bad AccessToken. see bruno but, the access token should not be null form
// {
//   "AccessToken": {
//     "access": null,
//     "manage": "",
//     "value": ""
//   },
//   "Continue": {
//     "access_token": {
//       "value": "5D91EE2CB6A64A718E7A"
//     },
//     "uri": "http://localhost:4006/continue/4c9d0eab-d9bd-4ac9-b4da-e02266e682ed"
//   }
// }

// TODO: use context. commented out for now for backwards compatibility
func (gs *GrantService) Request( url string, requestBody as.PostRequestJSONBody) (Grant2, error) {
// func (gs *GrantService) Request(ctx context.Context, url string, requestBody as.PostRequestJSONBody) (Grant2, error) {
	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Grant2{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	// req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return Grant2{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := gs.DoSigned(req)
	if err != nil {
		return Grant2{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Grant2{}, fmt.Errorf("failed to perform grant request: %s", resp.Status)
	}

	var grantResponse Grant2
	err = json.NewDecoder(resp.Body).Decode(&grantResponse)
	if err != nil {
		return Grant2{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return grantResponse, nil
}
