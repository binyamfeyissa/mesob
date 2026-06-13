package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type IdentityHTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (c *IdentityHTTPClient) RegisterCustomer(ctx context.Context, msisdn, lang string) error {
	type reqBody struct {
		MSISDN string `json:"msisdn"`
		Lang   string `json:"lang"`
	}
	body, _ := json.Marshal(reqBody{MSISDN: msisdn, Lang: lang})
	url := fmt.Sprintf("%s/identity/register", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("identity: build register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("identity: register customer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("identity: register returned status %d", resp.StatusCode)
	}
	return nil
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
