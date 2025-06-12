package openpayments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	as "github.com/interledger/open-payments-go-sdk/pkg/generated/authserver"
)

type TokenService struct {
	DoSigned RequestDoer
}

type TokenRotateParams struct {
	URL         string // The full URL of the token resource to rotate
	AccessToken string
}

type TokenRevokeParams struct {
	URL         string // The full URL of the token resource to revoke
	AccessToken string
}

func (ts *TokenService) Rotate(ctx context.Context, params TokenRotateParams) (as.AccessToken, error) {
	if params.URL == "" {
		return as.AccessToken{}, fmt.Errorf("missing required url")
	}
	if !strings.Contains(params.URL, "token/") {
		return as.AccessToken{}, fmt.Errorf("invalid token URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, params.URL, nil)
	if err != nil {
		return as.AccessToken{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := ts.DoSigned(req)
	if err != nil {
		return as.AccessToken{}, fmt.Errorf("rotate request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return as.AccessToken{}, fmt.Errorf("rotate request failed with status %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var response struct {
		AccessToken as.AccessToken `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return as.AccessToken{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.AccessToken, nil
}

func (ts *TokenService) Revoke(ctx context.Context, params TokenRevokeParams) error {
	if params.URL == "" || params.AccessToken == "" {
		return fmt.Errorf("missing required url or access token")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, params.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("GNAP %s", params.AccessToken))

	resp, err := ts.DoSigned(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

