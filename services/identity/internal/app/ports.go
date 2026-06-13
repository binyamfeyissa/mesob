package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/identity/internal/domain"
)

type UserRepository interface {
	Save(ctx context.Context, u *domain.User) error
	FindByMSISDN(ctx context.Context, msisdn string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateTier(ctx context.Context, id uuid.UUID, tier int8, version int) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error
	UpdateLanguage(ctx context.Context, id uuid.UUID, lang string) error
}

type CredentialRepository interface {
	Save(ctx context.Context, c *domain.Credential) error
	FindByUserID(ctx context.Context, id uuid.UUID) (*domain.Credential, error)
	IncrementFailed(ctx context.Context, id uuid.UUID) error
	Lock(ctx context.Context, id uuid.UUID, until time.Time) error
	ResetFailed(ctx context.Context, id uuid.UUID) error
	UpdatePINHash(ctx context.Context, id uuid.UUID, hash []byte) error
}

type KYCLimitsRepository interface {
	FindByTier(ctx context.Context, tier int8) (*domain.KYCLimits, error)
}

type OTPService interface {
	Send(ctx context.Context, msisdn, lang, channel string) (registrationID string, err error)
	Verify(ctx context.Context, registrationID, otp string) (challengeToken string, err error)
}

type NIDAdapter interface {
	VerifyFAN(ctx context.Context, fan, name, dob string) (verified bool, matchScore float64, err error)
}

type LedgerClient interface {
	CreateAccount(ctx context.Context, ownerType, ownerID, acctType, currency string) (accountID uuid.UUID, err error)
}

type SessionStore interface {
	CreateSession(ctx context.Context, userID uuid.UUID, role string) (accessToken, refreshToken string, err error)
	RevokeFamily(ctx context.Context, refreshToken string) error
	Rotate(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error)
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload any) error
}
