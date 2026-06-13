package app

import (
	"context"

	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type VerifyOTPInput struct {
	RegistrationID string
	OTP            string
}

type VerifyOTPOutput struct {
	Verified       bool
	Next           string
	ChallengeToken string
}

type VerifyOTPUseCase struct {
	OTP OTPService
}

func (uc *VerifyOTPUseCase) Execute(ctx context.Context, in VerifyOTPInput) (*VerifyOTPOutput, error) {
	// 1. Verify OTP — this deletes the Redis key on success (single-use).
	challengeToken, err := uc.OTP.Verify(ctx, in.RegistrationID, in.OTP)
	if err != nil {
		return nil, &kiterr.DomainError{
			Code:    "OTP_INVALID",
			Message: "incorrect or expired OTP",
		}
	}

	// 2. Return challenge token the client must present to SetPIN.
	return &VerifyOTPOutput{
		Verified:       true,
		Next:           "set_pin",
		ChallengeToken: challengeToken,
	}, nil
}
