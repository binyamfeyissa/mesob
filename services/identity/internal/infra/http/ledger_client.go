package httpinfra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
)

// LedgerHTTPClient implements app.LedgerClient by calling the Ledger service over HTTP.
type LedgerHTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type createAccountRequest struct {
	OwnerType string `json:"owner_type"`
	OwnerID   string `json:"owner_id"`
	AcctType  string `json:"acct_type"`
	Currency  string `json:"currency"`
}

type createAccountResponse struct {
	Data struct {
		AccountID string `json:"account_id"`
	} `json:"data"`
}

// CreateAccount posts to {BaseURL}/ledger/accounts and returns the new account UUID.
func (c *LedgerHTTPClient) CreateAccount(
	ctx context.Context,
	ownerType, ownerID, acctType, currency string,
) (uuid.UUID, error) {
	reqBody := createAccountRequest{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		AcctType:  acctType,
		Currency:  currency,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return uuid.Nil, err
	}

	url := c.BaseURL + "/ledger/accounts"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return uuid.Nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ledger create account: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return uuid.Nil, fmt.Errorf("ledger create account: unexpected status %d", resp.StatusCode)
	}

	var result createAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return uuid.Nil, fmt.Errorf("ledger create account: decode response: %w", err)
	}

	accountID, err := uuid.FromString(result.Data.AccountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ledger create account: invalid account_id %q: %w", result.Data.AccountID, err)
	}

	return accountID, nil
}
