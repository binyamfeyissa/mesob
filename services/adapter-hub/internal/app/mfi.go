package app

import (
	"context"

	"github.com/mesob-wallet/adapter-hub/internal/domain"
)

type MFIOriginateUseCase struct {
	Provider MFIProvider
	Mode     domain.AdapterMode
}

type MFIOriginateInput struct {
	UserRef        string
	PrincipalMinor int64
	TermDays       int
	ScoreRef       string
}

type MFIOriginateOutput struct {
	FacilityID string
	Status     string
	Mode       domain.AdapterMode
}

func (uc *MFIOriginateUseCase) Execute(ctx context.Context, in MFIOriginateInput) (MFIOriginateOutput, error) {
	if uc.Mode == domain.ModeDemo {
		return MFIOriginateOutput{FacilityID: "DEMO-" + in.UserRef, Status: "DISBURSED", Mode: domain.ModeDemo}, nil
	}
	fid, err := uc.Provider.Originate(ctx, in.UserRef, in.PrincipalMinor, in.TermDays, in.ScoreRef)
	if err != nil {
		return MFIOriginateOutput{}, err
	}
	return MFIOriginateOutput{FacilityID: fid, Status: "DISBURSED", Mode: domain.ModeLive}, nil
}
