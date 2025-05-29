package openpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	as "github.com/interledger/open-payments-go-sdk/pkg/generated/authserver"
)

// TODO: Improve cumbersome work of forming grant requests. See tests.
// - make new NewIncomingPaymentGrantRequest, etc. methods which
//   just have parmams for access items (and what else? just interact i guess?).
//   This means the methods need to be quite a bit smarter about grant requests but
//   that seems like the core purpose of the sdk - abstract away the complexity of sending
//   the request.
// - otherwise make helpers? ie for making Incoming Payment access token.
//   NewIncomingPaymentAccessToken that takes access items (anything else? maybe not)

type GrantService struct {
	DoSigned            RequestDoer
	client              string
}

// Same as generated as.PostRequestJSONBody, but without client.
// No way to copy type excluding the client field. This is
// prone to drift. Must update this if spec changes.
type GrantRequestBody struct {
	AccessToken struct {
		Access as.Access `json:"access"`
	} `json:"access_token"`

	Interact *as.InteractRequest `json:"interact,omitempty"`
}


// TODO: Address missing grant request type in generated types.
// This re-constructs from the generated types therefore is prone 
// to drift from OpenAPI spec. 
type Grant struct {
	Interact *as.InteractResponse `json:"interact,omitempty"`
	AccessToken *as.AccessToken `json:"access_token,omitempty"`
	Continue as.Continue `json:"continue"`
}

func (gr *Grant) IsInteractive() bool {
	return gr.Interact != nil
}

func (gr *Grant) IsGranted() bool {
	return gr.AccessToken != nil
}

func (gs *GrantService) Request(ctx context.Context, url string, requestBody as.PostRequestJSONBody) (Grant, error) {
	requestBody.Client = gs.client

	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Grant{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return Grant{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := gs.DoSigned(req)
	if err != nil {
		return Grant{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Grant{}, fmt.Errorf("failed to perform grant request: %s", resp.Status)
	}

	var grantResponse Grant
	err = json.NewDecoder(resp.Body).Decode(&grantResponse)
	if err != nil {
		return Grant{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return grantResponse, nil
}
