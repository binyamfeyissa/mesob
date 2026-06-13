package httpinfra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mesob-wallet/loans/internal/app"
)

type LedgerHTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (c *LedgerHTTPClient) PostTransaction(ctx context.Context, txnType, idemKey, initiatedBy, channel string, entries []app.LedgerEntry) (string, error) {
	type entryPayload struct {
		AccountID   string `json:"account_id"`
		Direction   string `json:"direction"`
		AmountMinor int64  `json:"amount_minor"`
	}
	payload := struct {
		Type        string         `json:"type"`
		IdemKey     string         `json:"idempotency_key"`
		InitiatedBy string         `json:"initiated_by"`
		Channel     string         `json:"channel"`
		Entries     []entryPayload `json:"entries"`
	}{
		Type: txnType, IdemKey: idemKey, InitiatedBy: initiatedBy, Channel: channel,
	}
	for _, e := range entries {
		payload.Entries = append(payload.Entries, entryPayload{
			AccountID: e.AccountID, Direction: e.Direction, AmountMinor: e.AmountMinor,
		})
	}

	b, err := json.Marshal(payload)
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
