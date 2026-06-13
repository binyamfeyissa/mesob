package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofrs/uuid"
)

type Claims struct {
	jwt.RegisteredClaims
	Role      string    `json:"role"`
	Scope     string    `json:"scope"`
	KYCTier   int8      `json:"kyc_tier"`
	RegionID  uuid.UUID `json:"region_id"`
	JTI       string    `json:"jti"` // unique token ID; used for refresh-family reuse detection
}

type contextKey string

const claimsKey contextKey = "claims"

func VerifyRS256(tokenString string, pubKey *rsa.PublicKey) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*Claims)
	return c, ok
}

func ContextWithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

func BearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	return strings.TrimPrefix(h, "Bearer "), true
}

type Role string

const (
	RoleEndUser       Role = "END_USER"
	RoleGroupLeader   Role = "GROUP_LEADER"
	RoleAgent         Role = "AGENT"
	RoleBranchOfficer Role = "BRANCH_OFFICER"
	RoleSuperAdmin    Role = "SUPER_ADMIN"
)

func HasRole(claims *Claims, required Role) bool {
	return claims.Role == string(required)
}

func IssueChallenge(userID uuid.UUID, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"type": "challenge",
		"exp":  time.Now().Add(10 * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}
