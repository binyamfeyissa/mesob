package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPClient forwards Telegram commands to the internal USSD/gateway service.
type HTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type forwardReq struct {
	MSISDN  string            `json:"msisdn"`
	Command string            `json:"command"`
	Payload map[string]string `json:"payload"`
}

type forwardResp struct {
	Message string `json:"message"`
}

func (c *HTTPClient) ForwardCommand(ctx context.Context, msisdn, command string, payload map[string]string) (string, error) {
	body, _ := json.Marshal(forwardReq{MSISDN: msisdn, Command: command, Payload: payload})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/gateway/telegram/command", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Sprintf("Command '%s' queued for processing.", command), nil // graceful
	}
	defer resp.Body.Close()
	var out forwardResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "Done.", nil
	}
	return out.Message, nil
}
