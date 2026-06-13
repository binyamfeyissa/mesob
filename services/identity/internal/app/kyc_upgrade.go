package app

import (
	"context"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type KYCUpgradeInput struct {
	UserID     uuid.UUID
	FAN        string
	FullName   string
	DOB        string
	TargetTier int8
}

type KYCUpgradeOutput struct {
	KYCTier int8
	Limits  struct {
		PerTxnMinor int64 `json:"per_txn_minor"`
		DailyMinor  int64 `json:"daily_minor"`
	}
}

type KYCUpgradeUseCase struct {
	Users      UserRepository
	NIDAdapter NIDAdapter
	Limits     KYCLimitsRepository
	Events     EventPublisher
}

func (uc *KYCUpgradeUseCase) Execute(ctx context.Context, in KYCUpgradeInput) (*KYCUpgradeOutput, error) {
	// 1. Load user.
	user, err := uc.Users.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "NOT_FOUND", Message: "user not found"}
	}

	// 2. Verify National ID (FAN) when adapter is available.
	if uc.NIDAdapter != nil {
		verified, _, err := uc.NIDAdapter.VerifyFAN(ctx, in.FAN, in.FullName, in.DOB)
		if err != nil || !verified {
			return nil, &kiterr.DomainError{
				Code:    "KYC_REJECTED",
				Message: "NID verification failed",
			}
		}
	}
	// If NIDAdapter is nil we proceed (development / stub mode).

	// 3. Upgrade tier on the domain object.
	user.UpgradeTier(in.TargetTier)

	// 4. Persist the updated user.
	if err := uc.Users.Save(ctx, user); err != nil {
		return nil, err
	}

	// 5. Publish KycTierChanged event.
	if uc.Events != nil {
		_ = uc.Events.Publish(ctx, "KycTierChanged", user.ID.String(), map[string]any{
			"kyc_tier": user.KYCTier,
		})
	}

	out := &KYCUpgradeOutput{KYCTier: user.KYCTier}

	// 6. Fetch tier limits when repository is available.
	if uc.Limits != nil {
		limits, err := uc.Limits.FindByTier(ctx, user.KYCTier)
		if err == nil && limits != nil {
			out.Limits.PerTxnMinor = limits.PerTxnMinor
			out.Limits.DailyMinor = limits.DailyMinor
		}
	} else {
		// Sensible defaults when Limits repo is not wired.
		out.Limits.PerTxnMinor = 5_000_000
		out.Limits.DailyMinor = 10_000_000
	}

	return out, nil
}
