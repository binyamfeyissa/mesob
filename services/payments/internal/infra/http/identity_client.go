package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type IdentityHTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type identityUserResponse struct {
	Data struct {
		UserID    string `json:"user_id"`
		AccountID string `json:"account_id"`
	} `json:"data"`
}

func (c *IdentityHTTPClient) FindUserByMSISDN(ctx context.Context, msisdn string) (userID string, accountID string, err error) {
	url := fmt.Sprintf("%s/identity/users/by-msisdn?msisdn=%s", c.BaseURL, msisdn)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("identity: build request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("identity: find user by msisdn: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", "", kiterr.ErrNotFound
	}
	if resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("identity: unexpected status %d", resp.StatusCode)
	}

	var result identityUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("identity: decode response: %w", err)
	}
	return result.Data.UserID, result.Data.AccountID, nil
}
