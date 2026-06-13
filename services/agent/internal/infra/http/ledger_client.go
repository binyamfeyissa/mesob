package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mesob-wallet/agent/internal/app"
)

type LedgerHTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type ledgerTxnRequest struct {
	Type        string            `json:"type"`
	InitiatedBy string            `json:"initiated_by"`
	Channel     string            `json:"channel"`
	Entries     []ledgerEntryJSON `json:"entries"`
}

type ledgerEntryJSON struct {
	AccountID   string `json:"account_id"`
	Direction   string `json:"direction"`
	AmountMinor int64  `json:"amount_minor"`
}

type ledgerTxnResponse struct {
	Data struct {
		TransactionID string `json:"transaction_id"`
	} `json:"data"`
}

type ledgerBalanceResponse struct {
	Data struct {
		BalanceMinor int64 `json:"balance_minor"`
	} `json:"data"`
}

func (c *LedgerHTTPClient) PostTransaction(
	ctx context.Context,
	txnType, idemKey, initiatedBy, channel string,
	entries []app.LedgerEntry,
) (string, error) {
	jsonEntries := make([]ledgerEntryJSON, len(entries))
	for i, e := range entries {
		jsonEntries[i] = ledgerEntryJSON{
			AccountID:   e.AccountID,
			Direction:   e.Direction,
			AmountMinor: e.AmountMinor,
		}
	}

	body, err := json.Marshal(ledgerTxnRequest{
		Type:        txnType,
		InitiatedBy: initiatedBy,
		Channel:     channel,
		Entries:     jsonEntries,
	})
	if err != nil {
		return "", fmt.Errorf("ledger: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/ledger/transactions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ledger: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idemKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ledger: post transaction: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("ledger: unexpected status %d", resp.StatusCode)
	}

	var result ledgerTxnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ledger: decode response: %w", err)
	}
	return result.Data.TransactionID, nil
}

func (c *LedgerHTTPClient) GetBalance(ctx context.Context, accountID string) (int64, error) {
	url := fmt.Sprintf("%s/ledger/accounts/%s/balance", c.BaseURL, accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("ledger: build balance request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("ledger: get balance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("ledger: unexpected status %d", resp.StatusCode)
	}

	var result ledgerBalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("ledger: decode balance response: %w", err)
	}
	return result.Data.BalanceMinor, nil
}
