package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
)

type sessionData struct {
	UserID  string `json:"user_id"`
	Role    string `json:"role"`
	KYCTier int8   `json:"kyc_tier"`
}

type SessionStore struct {
	Redis  *redis.Client
	Secret []byte
}

func (s *SessionStore) CreateSession(ctx context.Context, userID uuid.UUID, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = s.issueAccessToken(userID, role)
	if err != nil {
		return "", "", err
	}

	rtID, err := uuid.NewV4()
	if err != nil {
		return "", "", err
	}
	refreshToken = rtID.String()

	data, _ := json.Marshal(sessionData{UserID: userID.String(), Role: role})
	if err = s.Redis.Set(ctx, rtKey(refreshToken), data, refreshTTL).Err(); err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (s *SessionStore) Rotate(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error) {
	raw, err := s.Redis.GetDel(ctx, rtKey(refreshToken)).Result()
	if err != nil {
		return "", "", fmt.Errorf("invalid or expired refresh token")
	}

	var sd sessionData
	if err = json.Unmarshal([]byte(raw), &sd); err != nil {
		return "", "", fmt.Errorf("corrupt session data")
	}

	uid, err := uuid.FromString(sd.UserID)
	if err != nil {
		return "", "", fmt.Errorf("corrupt user id")
	}

	newAccess, err = s.issueAccessToken(uid, sd.Role)
	if err != nil {
		return "", "", err
	}

	rtID, _ := uuid.NewV4()
	newRefresh = rtID.String()
	data, _ := json.Marshal(sd)
	s.Redis.Set(ctx, rtKey(newRefresh), data, refreshTTL)
	return newAccess, newRefresh, nil
}

func (s *SessionStore) RevokeFamily(ctx context.Context, refreshToken string) error {
	return s.Redis.Del(ctx, rtKey(refreshToken)).Err()
}

func (s *SessionStore) issueAccessToken(userID uuid.UUID, role string) (string, error) {
	claims := gojwt.MapClaims{
		"sub":  userID.String(),
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(accessTTL).Unix(),
	}
	t := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return t.SignedString(s.Secret)
}

func rtKey(token string) string {
	return "rt:" + token
}
