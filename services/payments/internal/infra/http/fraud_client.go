package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type FraudHTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type fraudScreenRequest struct {
	UserID       string `json:"user_id"`
	Type         string `json:"type"`
	AmountMinor  int64  `json:"amount_minor"`
	Counterparty string `json:"counterparty"`
	Channel      string `json:"channel"`
}

type fraudScreenResponse struct {
	Data struct {
		Decision  string   `json:"decision"`
		RiskScore float64  `json:"risk_score"`
		RulesHit  []string `json:"rules_hit"`
	} `json:"data"`
}

func (c *FraudHTTPClient) Screen(
	ctx context.Context,
	userID, txnType string,
	amountMinor int64,
	counterparty, channel string,
) (decision string, riskScore float64, rulesHit []string, err error) {
	body, err := json.Marshal(fraudScreenRequest{
		UserID:       userID,
		Type:         txnType,
		AmountMinor:  amountMinor,
		Counterparty: counterparty,
		Channel:      channel,
	})
	if err != nil {
		return "", 0, nil, fmt.Errorf("fraud: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/fraud/screen", bytes.NewReader(body))
	if err != nil {
		return "", 0, nil, fmt.Errorf("fraud: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", 0, nil, fmt.Errorf("fraud: screen: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", 0, nil, fmt.Errorf("fraud: unexpected status %d", resp.StatusCode)
	}

	var result fraudScreenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, nil, fmt.Errorf("fraud: decode response: %w", err)
	}
	return result.Data.Decision, result.Data.RiskScore, result.Data.RulesHit, nil
}
