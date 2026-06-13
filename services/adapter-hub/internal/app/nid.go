package app

import (
	"context"

	"github.com/mesob-wallet/adapter-hub/internal/domain"
)

type NIDVerifyUseCase struct {
	Provider NIDProvider
	Mode     domain.AdapterMode
}

type NIDVerifyInput struct {
	FAN  string
	Name string
	DOB  string
}

type NIDVerifyOutput struct {
	Verified   bool
	MatchScore float64
	Mode       domain.AdapterMode
}

func (uc *NIDVerifyUseCase) Execute(ctx context.Context, in NIDVerifyInput) (NIDVerifyOutput, error) {
	if uc.Mode == domain.ModeDemo {
		return NIDVerifyOutput{Verified: true, MatchScore: 0.95, Mode: domain.ModeDemo}, nil
	}
	verified, score, err := uc.Provider.Verify(ctx, in.FAN, in.Name, in.DOB)
	if err != nil {
		return NIDVerifyOutput{}, err
	}
	return NIDVerifyOutput{Verified: verified, MatchScore: score, Mode: domain.ModeLive}, nil
}
