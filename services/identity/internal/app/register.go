package app

import (
	"context"
	"regexp"
	"strings"

	kiterr "github.com/mesob-wallet/go-kit/errors"
)

// ethiopianLocal matches local Ethiopian format: 09XXXXXXXX or 07XXXXXXXX (10 digits total).
var ethiopianLocal = regexp.MustCompile(`^0([79]\d{8})$`)

// ethiopianE164 matches E.164 Ethiopian format: +2519XXXXXXXX or +2517XXXXXXXX.
var ethiopianE164 = regexp.MustCompile(`^\+251([79]\d{8})$`)

// normalizeMSISDN normalises an Ethiopian mobile number to E.164 (+251XXXXXXXXX).
// Returns ("", false) if the number is not a valid Ethiopian mobile number.
func normalizeMSISDN(msisdn string) (string, bool) {
	msisdn = strings.TrimSpace(msisdn)
	if m := ethiopianE164.FindStringSubmatch(msisdn); m != nil {
		return "+251" + m[1], true
	}
	if m := ethiopianLocal.FindStringSubmatch(msisdn); m != nil {
		return "+251" + m[1], true
	}
	return "", false
}

type RegisterInput struct {
	MSISDN   string
	RegionID string
	Lang     string
}

type RegisterOutput struct {
	RegistrationID string
	OTPChannel     string
	ExpiresIn      int
}

type RegisterUseCase struct {
	Users UserRepository
	OTP   OTPService
}

func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (*RegisterOutput, error) {
	// 1. Validate and normalise Ethiopian MSISDN.
	normalized, ok := normalizeMSISDN(in.MSISDN)
	if !ok {
		return nil, &kiterr.DomainError{
			Code:    "INVALID_MSISDN",
			Message: "MSISDN must be a valid Ethiopian mobile number (+251 or 0 followed by 9 digits)",
		}
	}
	in.MSISDN = normalized

	// 2. Check whether an active account already exists for this MSISDN.
	existing, err := uc.Users.FindByMSISDN(ctx, in.MSISDN)
	if err == nil && existing != nil {
		if existing.IsActive() {
			return nil, &kiterr.DomainError{
				Code:    "MSISDN_TAKEN",
				Message: "an active account already exists for this phone number",
			}
		}
	}

	// 3. Send OTP.
	regID, err := uc.OTP.Send(ctx, in.MSISDN, in.Lang, "SMS")
	if err != nil {
		return nil, err
	}

	// 4. Return registration details.
	return &RegisterOutput{
		RegistrationID: regID,
		OTPChannel:     "SMS",
		ExpiresIn:      300,
	}, nil
}
