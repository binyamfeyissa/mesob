package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	gojwt "github.com/golang-jwt/jwt/v5"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	kitresp "github.com/mesob-wallet/go-kit/response"
	"github.com/mesob-wallet/identity/internal/app"
	"github.com/mesob-wallet/identity/internal/domain"
)

type Handler struct {
	Register   *app.RegisterUseCase
	VerifyOTP  *app.VerifyOTPUseCase
	SetPIN     *app.SetPINUseCase
	Login      *app.LoginUseCase
	KYCUpgrade *app.KYCUpgradeUseCase
	Users      app.UserRepository
	JWTSecret  []byte
}

func (h *Handler) Register_(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MSISDN   string `json:"msisdn"`
		RegionID string `json:"region_id"`
		Lang     string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	out, err := h.Register.Execute(r.Context(), app.RegisterInput{
		MSISDN:   req.MSISDN,
		RegionID: req.RegionID,
		Lang:     req.Lang,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusAccepted, out, meta(r))
}

func (h *Handler) VerifyOTP_(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RegistrationID string `json:"registration_id"`
		OTP            string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	out, err := h.VerifyOTP.Execute(r.Context(), app.VerifyOTPInput{
		RegistrationID: req.RegistrationID,
		OTP:            req.OTP,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) Login_(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MSISDN string `json:"msisdn"`
		PIN    string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	out, err := h.Login.Execute(r.Context(), app.LoginInput{
		MSISDN: req.MSISDN,
		PIN:    req.PIN,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "rt",
		Value:    out.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})

	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) TokenRefresh(w http.ResponseWriter, r *http.Request) {
	rt := ""
	if cookie, err := r.Cookie("rt"); err == nil {
		rt = cookie.Value
	}
	if rt == "" {
		rt = r.Header.Get("X-Refresh-Token")
	}
	if rt == "" {
		kitresp.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "refresh token missing", meta(r))
		return
	}

	newAccess, newRefresh, err := h.Login.Sessions.Rotate(r.Context(), rt)
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name: "rt", Value: "", Path: "/",
			HttpOnly: true, MaxAge: -1, Expires: time.Unix(0, 0),
		})
		kitresp.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "session expired", meta(r))
		return
	}

	// Extract user info from new access token claims
	var userID, role string
	var kycTier int8
	var walletAccountID string
	parsed := gojwt.MapClaims{}
	if _, parseErr := gojwt.ParseWithClaims(newAccess, parsed, func(t *gojwt.Token) (interface{}, error) {
		return h.JWTSecret, nil
	}); parseErr == nil {
		userID, _ = parsed["sub"].(string)
		role, _ = parsed["role"].(string)
		if h.Users != nil {
			if uid, err2 := uuid.FromString(userID); err2 == nil {
				if u, err2 := h.Users.FindByID(r.Context(), uid); err2 == nil {
					kycTier = u.KYCTier
					if u.WalletAccountID != nil {
						walletAccountID = u.WalletAccountID.String()
					}
				}
			}
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "rt",
		Value:    newRefresh,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})

	kitresp.Success(w, http.StatusOK, map[string]any{
		"access_token":     newAccess,
		"refresh_token":    newRefresh,
		"expires_in":       900,
		"user_id":          userID,
		"role":             role,
		"kyc_tier":         kycTier,
		"wallet_account_id": walletAccountID,
	}, meta(r))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	rt := ""
	if cookie, err := r.Cookie("rt"); err == nil {
		rt = cookie.Value
	}
	if rt == "" {
		rt = r.Header.Get("X-Refresh-Token")
	}
	if rt != "" && h.Login.Sessions != nil {
		_ = h.Login.Sessions.RevokeFamily(r.Context(), rt)
	}
	http.SetCookie(w, &http.Cookie{
		Name: "rt", Value: "", Path: "/",
		HttpOnly: true, MaxAge: -1, Expires: time.Unix(0, 0),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		kitresp.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token", meta(r))
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims := gojwt.MapClaims{}
	_, err := gojwt.ParseWithClaims(tokenStr, claims, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return h.JWTSecret, nil
	})
	if err != nil {
		kitresp.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token", meta(r))
		return
	}

	userIDStr, _ := claims["sub"].(string)
	uid, err := uuid.FromString(userIDStr)
	if err != nil {
		kitresp.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user id in token", meta(r))
		return
	}

	if h.Users == nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", "user repository not wired", meta(r))
		return
	}

	user, err := h.Users.FindByID(r.Context(), uid)
	if err != nil {
		kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "user not found", meta(r))
		return
	}

	kitresp.Success(w, http.StatusOK, map[string]any{
		"user_id":  user.ID.String(),
		"msisdn":   user.MSISDN,
		"role":     user.Role,
		"kyc_tier": user.KYCTier,
		"status":   string(user.Status),
		"lang":     user.PreferredLang,
	}, meta(r))
}

func (h *Handler) PatchLanguage(w http.ResponseWriter, r *http.Request) {
	userID := subFromBearer(r)
	var req struct {
		Lang string `json:"lang"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Lang == "" {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", "lang required", meta(r))
		return
	}
	if err := h.Users.UpdateLanguage(r.Context(), userID, req.Lang); err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]string{"lang": req.Lang}, meta(r))
}

func (h *Handler) KYCUpgrade_(w http.ResponseWriter, r *http.Request) {
	userID := subFromBearer(r)
	var req struct {
		FAN        string `json:"fan"`
		FullName   string `json:"full_name"`
		DOB        string `json:"dob"`
		TargetTier int8   `json:"target_tier"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	out, err := h.KYCUpgrade.Execute(r.Context(), app.KYCUpgradeInput{
		UserID:     userID,
		FAN:        req.FAN,
		FullName:   req.FullName,
		DOB:        req.DOB,
		TargetTier: req.TargetTier,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) ChangePIN(w http.ResponseWriter, r *http.Request) {
	userID := subFromBearer(r)
	var req struct {
		OldPIN string `json:"old_pin"`
		NewPIN string `json:"new_pin"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	cred, err := h.Login.Creds.FindByUserID(r.Context(), userID)
	if err != nil {
		kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "credential not found", meta(r))
		return
	}
	if !cred.VerifyPIN(req.OldPIN) {
		kitresp.Err(w, http.StatusUnauthorized, "INVALID_PIN", "old PIN incorrect", meta(r))
		return
	}
	newHash, err := domain.HashPIN(req.NewPIN)
	if err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	if err := h.Login.Creds.UpdatePINHash(r.Context(), userID, newHash); err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetPIN_stub(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChallengeToken string `json:"challenge_token"`
		MSISDN         string `json:"msisdn"`
		PIN            string `json:"pin"`
		RegionID       string `json:"region_id"`
		Lang           string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	regionID, _ := uuid.FromString(req.RegionID)
	out, err := h.SetPIN.Execute(r.Context(), app.SetPINInput{
		ChallengeToken: req.ChallengeToken,
		MSISDN:         req.MSISDN,
		PIN:            req.PIN,
		RegionID:       regionID,
		Lang:           req.Lang,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

// GetUserByMSISDN is an internal service-to-service endpoint used by the agent
// service to resolve a user's ID and ledger wallet account ID by their MSISDN.
// It does NOT require a user JWT — callers must be on the internal network.
func (h *Handler) GetUserByMSISDN(w http.ResponseWriter, r *http.Request) {
	msisdn := r.URL.Query().Get("msisdn")
	if msisdn == "" {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", "msisdn query param required", meta(r))
		return
	}
	if h.Users == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "NOT_READY", "user repo not wired", meta(r))
		return
	}
	user, err := h.Users.FindByMSISDN(r.Context(), msisdn)
	if err != nil {
		kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "user not found", meta(r))
		return
	}
	accountID := ""
	if user.WalletAccountID != nil {
		accountID = user.WalletAccountID.String()
	}
	kitresp.Success(w, http.StatusOK, map[string]any{
		"user_id":    user.ID.String(),
		"account_id": accountID,
	}, meta(r))
}

// GetUserByID is an internal service-to-service endpoint used by the notification
// service to resolve a user's contact info (MSISDN, language) by their UUID.
// It does NOT require a user JWT — callers must be on the internal network.
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	uid, err := uuid.FromString(idStr)
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid user id", meta(r))
		return
	}
	if h.Users == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "NOT_READY", "user repo not wired", meta(r))
		return
	}
	user, err := h.Users.FindByID(r.Context(), uid)
	if err != nil {
		kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "user not found", meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]any{
		"user_id":          user.ID.String(),
		"msisdn":           user.MSISDN,
		"preferred_lang":   user.PreferredLang,
		"telegram_chat_id": int64(0), // populated when user links Telegram
		"fcm_token":        "",       // populated when user grants push permission
	}, meta(r))
}

// subFromBearer extracts the "sub" claim from a JWT Bearer token without full
// validation (signature check happens in middleware / GetMe). Returns uuid.Nil
// if the token is missing or malformed.
func subFromBearer(r *http.Request) uuid.UUID {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return uuid.Nil
	}
	parts := strings.Split(strings.TrimPrefix(auth, "Bearer "), ".")
	if len(parts) != 3 {
		return uuid.Nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.Nil
	}
	sub, _ := claims["sub"].(string)
	id, _ := uuid.FromString(sub)
	return id
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	if de, ok := err.(*kiterr.DomainError); ok {
		status := http.StatusUnprocessableEntity
		switch de.Code {
		case "NOT_FOUND":
			status = http.StatusNotFound
		case "INVALID_PIN", "ACCOUNT_LOCKED", "ACCOUNT_SUSPENDED":
			status = http.StatusUnauthorized
		case "FRAUD_BLOCKED":
			status = http.StatusForbidden
		case "IDEMPOTENCY_MISMATCH":
			status = http.StatusConflict
		}
		kitresp.Err(w, status, de.Code, de.Message, meta(r))
		return
	}
	kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
}

func meta(r *http.Request) kitresp.MetaBlock {
	return kitresp.MetaBlock{
		RequestID: r.Header.Get("X-Request-ID"),
		TraceID:   r.Header.Get("X-Trace-ID"),
	}
}
