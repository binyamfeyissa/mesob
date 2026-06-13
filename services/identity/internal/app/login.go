package app

import (
	"context"
	"time"

	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type LoginInput struct {
	MSISDN string
	PIN    string
}

type LoginOutput struct {
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	ExpiresIn       int    `json:"expires_in"`
	Role            string `json:"role"`
	KYCTier         int8   `json:"kyc_tier"`
	UserID          string `json:"user_id"`
	WalletAccountID string `json:"wallet_account_id,omitempty"`
}

type LoginUseCase struct {
	Users    UserRepository
	Creds    CredentialRepository
	Sessions SessionStore
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	user, err := uc.Users.FindByMSISDN(ctx, in.MSISDN)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "NOT_FOUND", Message: "user not found"}
	}
	if !user.IsActive() {
		return nil, &kiterr.DomainError{Code: "ACCOUNT_SUSPENDED", Message: "account is not active"}
	}

	cred, err := uc.Creds.FindByUserID(ctx, user.ID)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "INTERNAL", Message: "credential lookup failed"}
	}
	if cred.IsLocked() {
		return nil, &kiterr.DomainError{Code: "ACCOUNT_LOCKED", Message: "account locked — try again in 30 minutes"}
	}

	if !cred.VerifyPIN(in.PIN) {
		_ = uc.Creds.IncrementFailed(ctx, user.ID)
		cred.FailedCount++
		if cred.ShouldLock() {
			lockUntil := time.Now().Add(30 * time.Minute)
			_ = uc.Creds.Lock(ctx, user.ID, lockUntil)
		}
		return nil, &kiterr.DomainError{Code: "INVALID_PIN", Message: "incorrect PIN"}
	}

	_ = uc.Creds.ResetFailed(ctx, user.ID)

	accessToken, refreshToken, err := uc.Sessions.CreateSession(ctx, user.ID, user.Role)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "INTERNAL", Message: "session creation failed"}
	}

	walletAccountID := ""
	if user.WalletAccountID != nil {
		walletAccountID = user.WalletAccountID.String()
	}

	return &LoginOutput{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		ExpiresIn:       900,
		Role:            user.Role,
		KYCTier:         user.KYCTier,
		UserID:          user.ID.String(),
		WalletAccountID: walletAccountID,
	}, nil
}
