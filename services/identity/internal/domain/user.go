package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type UserStatus string

const (
	StatusPending   UserStatus = "PENDING"
	StatusActive    UserStatus = "ACTIVE"
	StatusLocked    UserStatus = "LOCKED"
	StatusSuspended UserStatus = "SUSPENDED"
)

type User struct {
	ID               uuid.UUID
	MSISDN           string
	FANCiphertext    []byte
	FANHash          []byte
	FullNameEnc      []byte
	KYCTier          int8
	Role             string
	RegionID         uuid.UUID
	Status           UserStatus
	PreferredLang    string
	WalletAccountID  *uuid.UUID
	Version          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func NewUser(msisdn string, regionID uuid.UUID, lang string) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &User{
		ID:            id,
		MSISDN:        msisdn,
		KYCTier:       0,
		RegionID:      regionID,
		Status:        StatusPending,
		PreferredLang: lang,
		Version:       1,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

func (u *User) Activate() {
	u.Status = StatusActive
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Lock() {
	u.Status = StatusLocked
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) UpgradeTier(tier int8) {
	u.KYCTier = tier
	u.Version++
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}
