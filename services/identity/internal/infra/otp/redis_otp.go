package otp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
)

const otpTTL = 300 * time.Second

func otpSigningSecret() string {
	if s := os.Getenv("MESOB_OTP_SECRET"); s != "" {
		return s
	}
	return "mesob-otp-secret"
}

type otpRecord struct {
	OTP    string `json:"otp"`
	MSISDN string `json:"msisdn"`
	Lang   string `json:"lang"`
}

// RedisOTP implements app.OTPService using Redis as the backing store.
type RedisOTP struct {
	Redis *redis.Client
}

// Send generates a 6-digit OTP, stores it in Redis with a 300 s TTL, and returns
// a registrationID that the client uses when calling Verify.
// No real SMS is sent — the OTP is printed to stdout for development.
func (o *RedisOTP) Send(ctx context.Context, msisdn, lang, channel string) (string, error) {
	otp := fmt.Sprintf("%06d", rand.Intn(1_000_000))

	registrationID := uuid.Must(uuid.NewV4()).String()

	rec := otpRecord{OTP: otp, MSISDN: msisdn, Lang: lang}
	payload, err := json.Marshal(rec)
	if err != nil {
		return "", err
	}

	key := "otp:" + registrationID
	if err := o.Redis.Set(ctx, key, payload, otpTTL).Err(); err != nil {
		return "", err
	}

	// Dev: print OTP to stdout so engineers can test without a real SMS gateway.
	fmt.Printf("[DEV] OTP for %s (registrationID=%s): %s\n", msisdn, registrationID, otp)

	return registrationID, nil
}

// Verify checks the supplied OTP against the stored record.
// On success it deletes the key (single-use) and returns a challengeToken that the
// caller must present to SetPIN to prove they passed OTP verification.
func (o *RedisOTP) Verify(ctx context.Context, registrationID, otp string) (string, error) {
	key := "otp:" + registrationID

	raw, err := o.Redis.Get(ctx, key).Bytes()
	if err != nil {
		return "", fmt.Errorf("OTP_INVALID: OTP not found or expired")
	}

	var rec otpRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return "", fmt.Errorf("OTP_INVALID: malformed OTP record")
	}

	if rec.OTP != otp {
		return "", fmt.Errorf("OTP_INVALID: incorrect OTP")
	}

	// Delete on success — single-use.
	_ = o.Redis.Del(ctx, key)

	challengeToken := deriveChallenge(registrationID, rec.MSISDN)
	return challengeToken, nil
}

// deriveChallenge computes HMAC-SHA256(registrationID + msisdn + secret) and returns
// the result hex-encoded.
func deriveChallenge(registrationID, msisdn string) string {
	mac := hmac.New(sha256.New, []byte(otpSigningSecret()))
	mac.Write([]byte(registrationID + msisdn))
	return hex.EncodeToString(mac.Sum(nil))
}
