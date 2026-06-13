package httpinfra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mesob-wallet/loans/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type ScoringHTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func (c *ScoringHTTPClient) Score(ctx context.Context, userID string, forceRecompute bool) (*domain.CreditScore, error) {
	payload := struct {
		UserID         string `json:"user_id"`
		ForceRecompute bool   `json:"force_recompute"`
	}{UserID: userID, ForceRecompute: forceRecompute}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/scoring/score", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, kiterr.ErrScoringDeferred
	}
	defer resp.Body.Close()

	if resp.StatusCode == 422 {
		return nil, kiterr.ErrInsufficientHistory
	}
	if resp.StatusCode >= 500 {
		return nil, kiterr.ErrScoringDeferred
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("scoring returned %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			ScoreID      string  `json:"score_id"`
			Score        int     `json:"score"`
			Tier         string  `json:"tier"`
			CeilingMinor int64   `json:"ceiling_minor"`
			ModelVer     string  `json:"model_ver"`
			Source       string  `json:"source"`
			Factors      []struct {
				Feature      string  `json:"feature"`
				Contribution float64 `json:"contribution"`
			} `json:"factors"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	cs := &domain.CreditScore{
		ScoreID:      result.Data.ScoreID,
		Score:        result.Data.Score,
		Tier:         result.Data.Tier,
		CeilingMinor: result.Data.CeilingMinor,
		ModelVer:     result.Data.ModelVer,
		Source:       result.Data.Source,
	}
	for _, f := range result.Data.Factors {
		cs.Factors = append(cs.Factors, domain.Factor{
			Feature:      f.Feature,
			Contribution: f.Contribution,
		})
	}
	return cs, nil
}
