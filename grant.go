package openpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	as "github.com/interledger/open-payments-go/generated/authserver"
)

// GrantRequestOption is a functional option for customizing a grant request.
type GrantRequestOption func(*as.GrantRequestWithAccessToken)

// WithInteract sets the interact field on the grant request.
func WithInteract(interact *as.InteractRequest) GrantRequestOption {
	return func(req *as.GrantRequestWithAccessToken) {
		req.Interact = interact
	}
}

// WithSubject sets the subject field on the grant request.
func WithSubject(subject *as.Subject) GrantRequestOption {
	return func(req *as.GrantRequestWithAccessToken) {
		req.Subject = subject
	}
}

// NewIncomingPaymentGrantRequestParams creates a GrantRequestParams for an incoming payment grant.
func NewIncomingPaymentGrantRequestParams(
	url string,
	actions []as.AccessIncomingActions,
	opts ...GrantRequestOption,
) (GrantRequestParams, error) {
	access := as.AccessIncoming{
		Type:    as.IncomingPayment,
		Actions: actions,
	}
	accessItem := as.AccessItem{}
	if err := accessItem.FromAccessIncoming(access); err != nil {
		return GrantRequestParams{}, fmt.Errorf("failed to create access item: %w", err)
	}

	body := as.GrantRequestWithAccessToken{
		AccessToken: as.AccessTokenRequest{
			Access: []as.AccessItem{accessItem},
		},
	}
	for _, opt := range opts {
		opt(&body)
	}

	return GrantRequestParams{URL: url, RequestBody: body}, nil
}

// NewQuoteGrantRequestParams creates a GrantRequestParams for a quote grant.
func NewQuoteGrantRequestParams(
	url string,
	actions []as.AccessQuoteActions,
	opts ...GrantRequestOption,
) (GrantRequestParams, error) {
	access := as.AccessQuote{
		Type:    as.Quote,
		Actions: actions,
	}
	accessItem := as.AccessItem{}
	if err := accessItem.FromAccessQuote(access); err != nil {
		return GrantRequestParams{}, fmt.Errorf("failed to create access item: %w", err)
	}

	body := as.GrantRequestWithAccessToken{
		AccessToken: as.AccessTokenRequest{
			Access: []as.AccessItem{accessItem},
		},
	}
	for _, opt := range opts {
		opt(&body)
	}

	return GrantRequestParams{URL: url, RequestBody: body}, nil
}

// NewOutgoingPaymentGrantRequestParams creates a GrantRequestParams for an outgoing payment grant.
func NewOutgoingPaymentGrantRequestParams(
	url string,
	identifier string,
	actions []as.AccessOutgoingActions,
	opts ...GrantRequestOption,
) (GrantRequestParams, error) {
	access := as.AccessOutgoing{
		Type:       as.OutgoingPayment,
		Actions:    actions,
		Identifier: identifier,
	}
	accessItem := as.AccessItem{}
	if err := accessItem.FromAccessOutgoing(access); err != nil {
		return GrantRequestParams{}, fmt.Errorf("failed to create access item: %w", err)
	}

	body := as.GrantRequestWithAccessToken{
		AccessToken: as.AccessTokenRequest{
			Access: []as.AccessItem{accessItem},
		},
	}
	for _, opt := range opts {
		opt(&body)
	}

	return GrantRequestParams{URL: url, RequestBody: body}, nil
}

type GrantService struct {
	DoSigned RequestDoer
	client   string
}

type GrantRequestParams struct {
	URL         string // Auth server URL
	RequestBody as.GrantRequestWithAccessToken
}

type GrantCancelParams struct {
	URL         string // continue URI
	AccessToken string
}

type GrantContinueParams struct {
	URL         string
	AccessToken string
	InteractRef string
}

// TODO: Address missing grant request type in generated types.
// This re-constructs from the generated types therefore is prone
// to drift from OpenAPI spec.
type Grant struct {
	Interact    *as.InteractResponse `json:"interact,omitempty"`
	AccessToken *as.AccessToken      `json:"access_token,omitempty"`
	Continue    as.Continue          `json:"continue"`
}

func (gr *Grant) IsInteractive() bool {
	return gr.Interact != nil
}

func (gr *Grant) IsGranted() bool {
	return gr.AccessToken != nil
}

func (gs *GrantService) Request(ctx context.Context, params GrantRequestParams) (Grant, error) {
	params.RequestBody.Client = gs.client

	reqBodyBytes, err := json.Marshal(params.RequestBody)
	if err != nil {
		return Grant{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// TODO: validation. [debit|receive]Amount limits are mutually exclusive for
	// access token's with type outgoing-payment

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, params.URL, bytes.NewBuffer(reqBodyBytes))
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
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		bodyStr := strings.TrimSpace(string(bodyBytes))

		return Grant{}, fmt.Errorf(
			"failed to perform grant request: %s\nURL: %s\nRequest body: %s\nResponse body: %s",
			resp.Status, req.URL, string(reqBodyBytes), bodyStr,
		)
	}

	var grantResponse Grant
	err = json.NewDecoder(resp.Body).Decode(&grantResponse)
	if err != nil {
		return Grant{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	return grantResponse, nil
}

func (gs *GrantService) Continue(ctx context.Context, params GrantContinueParams) (Grant, error) {
	if params.URL == "" || params.AccessToken == "" {
		return Grant{}, fmt.Errorf("missing required url or access token")
	}
	if !strings.Contains(params.URL, "continue/") {
		return Grant{}, fmt.Errorf("invalid continuation grant URL: %s", params.URL)
	}

	requestBody := map[string]string{}

	if (params.InteractRef != "") {
		requestBody["interact_ref"] = params.InteractRef
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return Grant{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, params.URL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return Grant{}, fmt.Errorf("failed to create continue request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GNAP "+params.AccessToken)

	resp, err := gs.DoSigned(req)
	if err != nil {
		return Grant{}, fmt.Errorf("continue request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Grant{}, fmt.Errorf("continue request failed with status %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var grantResponse Grant
	if err := json.NewDecoder(resp.Body).Decode(&grantResponse); err != nil {
		return Grant{}, fmt.Errorf("failed to decode continue response: %w", err)
	}

	return grantResponse, nil
}

func (gs *GrantService) Cancel(ctx context.Context, params GrantCancelParams) error {
	if params.URL == "" || params.AccessToken == "" {
		return fmt.Errorf("missing required url or access token")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, params.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create cancel request: %w", err)
	}
	req.Header.Set("Authorization", "GNAP "+params.AccessToken)

	resp, err := gs.DoSigned(req)
	if err != nil {
		return fmt.Errorf("cancel request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("cancel request failed with status %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}
