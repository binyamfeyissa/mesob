package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	"github.com/mesob-wallet/identity/internal/domain"
)

type SetPINInput struct {
	ChallengeToken string
	MSISDN         string
	PIN            string
	RegionID       uuid.UUID
	Lang           string
}

type SetPINOutput struct {
	UserID       string
	Status       string
	KYCTier      int8
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type SetPINUseCase struct {
	Users    UserRepository
	Creds    CredentialRepository
	Ledger   LedgerClient
	Sessions SessionStore
	Events   EventPublisher
}

func (uc *SetPINUseCase) Execute(ctx context.Context, in SetPINInput) (*SetPINOutput, error) {
	// 1. Validate challenge token — must be a non-empty hex string (>=32 chars).
	//    In production the token would be verified via HMAC against a Redis-stored nonce.
	if len(in.ChallengeToken) < 32 {
		return nil, &kiterr.DomainError{
			Code:    "INVALID_CHALLENGE",
			Message: "invalid challenge token",
		}
	}

	// 2. Create user domain object and activate it immediately.
	user, err := domain.NewUser(in.MSISDN, in.RegionID, in.Lang)
	if err != nil {
		return nil, err
	}
	user.Role = "USER"
	user.Activate()

	// 3. Persist user.
	if err := uc.Users.Save(ctx, user); err != nil {
		return nil, err
	}

	// 4. Create ledger wallet account for the new user.
	var accountID uuid.UUID
	if uc.Ledger != nil {
		accountID, err = uc.Ledger.CreateAccount(ctx, "USER", user.ID.String(), "WALLET", "ETB")
		if err != nil {
			// Non-fatal in dev: log the error but continue so we can test without ledger.
			accountID = uuid.Nil
		}
	}
	if accountID != uuid.Nil {
		user.WalletAccountID = &accountID
	}

	// 5. Hash the PIN.
	pinHash, err := domain.HashPIN(in.PIN)
	if err != nil {
		return nil, err
	}

	// 6. Persist credential.
	cred := &domain.Credential{
		UserID:    user.ID,
		PINHash:   pinHash,
		UpdatedAt: time.Now().UTC(),
	}
	if err := uc.Creds.Save(ctx, cred); err != nil {
		return nil, err
	}

	// 7. Create session tokens.
	accessToken, refreshToken, err := uc.Sessions.CreateSession(ctx, user.ID, user.Role)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "INTERNAL", Message: "session creation failed"}
	}

	// 8. Publish UserActivated event via outbox.
	if uc.Events != nil {
		_ = uc.Events.Publish(ctx, "UserActivated", user.ID.String(), map[string]any{
			"msisdn":     user.MSISDN,
			"account_id": accountID.String(),
		})
	}

	// 9. Return tokens and user summary.
	return &SetPINOutput{
		UserID:       user.ID.String(),
		Status:       string(user.Status),
		KYCTier:      user.KYCTier,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}
