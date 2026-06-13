package httpinfra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mesob-wallet/iqub/internal/app"
)

type LedgerHTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (c *LedgerHTTPClient) PostTransaction(ctx context.Context, txnType, idemKey, initiatedBy, channel string, entries []app.LedgerEntry) (string, error) {
	type entryJSON struct {
		AccountID   string `json:"account_id"`
		Direction   string `json:"direction"`
		AmountMinor int64  `json:"amount_minor"`
	}
	body := struct {
		Type           string      `json:"type"`
		IdempotencyKey string      `json:"idempotency_key"`
		InitiatedBy    string      `json:"initiated_by"`
		Channel        string      `json:"channel"`
		Entries        []entryJSON `json:"entries"`
	}{Type: txnType, IdempotencyKey: idemKey, InitiatedBy: initiatedBy, Channel: channel}
	for _, e := range entries {
		body.Entries = append(body.Entries, entryJSON{AccountID: e.AccountID, Direction: e.Direction, AmountMinor: e.AmountMinor})
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/ledger/transactions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ledger returned %d", resp.StatusCode)
	}
	var result struct {
		Data struct {
			TransactionID string `json:"transaction_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.TransactionID, nil
}

func (c *LedgerHTTPClient) CreateAccount(ctx context.Context, ownerType, ownerID, acctType, currency string) (string, error) {
	body := struct {
		OwnerType string `json:"owner_type"`
		OwnerID   string `json:"owner_id"`
		Type      string `json:"type"`
		Currency  string `json:"currency"`
	}{OwnerType: ownerType, OwnerID: ownerID, Type: acctType, Currency: currency}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/ledger/accounts", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ledger returned %d", resp.StatusCode)
	}
	var result struct {
		Data struct {
			AccountID string `json:"account_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.AccountID, nil
}
