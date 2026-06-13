package app

import (
	"context"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type EligibilityOutput struct {
	Eligible     bool   `json:"eligible"`
	Tier         string `json:"tier"`
	Score        int    `json:"score"`
	CeilingMinor int64  `json:"ceiling_minor"`
	Factors      []struct {
		Feature string  `json:"feature"`
		Impact  float64 `json:"impact"`
	} `json:"factors"`
}

type CheckEligibilityUseCase struct {
	Scoring ScoringClient
}

func (uc *CheckEligibilityUseCase) Execute(ctx context.Context, userID uuid.UUID) (*EligibilityOutput, error) {
	if uc.Scoring == nil {
		return &EligibilityOutput{Eligible: false, Tier: "UNSCORED", Score: 0, CeilingMinor: 0}, nil
	}
	cs, err := uc.Scoring.Score(ctx, userID.String(), false)
	if err != nil {
		if kiterr.Is(err, kiterr.ErrInsufficientHistory) {
			return nil, err
		}
		if kiterr.Is(err, kiterr.ErrScoringDeferred) {
			return nil, err
		}
		return nil, err
	}
	eligible := cs.Score >= 300 && cs.CeilingMinor > 0
	factors := make([]struct {
		Feature string  `json:"feature"`
		Impact  float64 `json:"impact"`
	}, len(cs.Factors))
	for i, f := range cs.Factors {
		factors[i] = struct {
			Feature string  `json:"feature"`
			Impact  float64 `json:"impact"`
		}{Feature: f.Feature, Impact: f.Contribution}
	}
	return &EligibilityOutput{
		Eligible:     eligible,
		Tier:         cs.Tier,
		Score:        cs.Score,
		CeilingMinor: cs.CeilingMinor,
		Factors:      factors,
	}, nil
}
