package http

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /identity/register",      h.Register_)
	mux.HandleFunc("POST /identity/verify-otp",    h.VerifyOTP_)
	mux.HandleFunc("POST /identity/set-pin",       h.SetPIN_stub)
	mux.HandleFunc("POST /identity/login",         h.Login_)
	mux.HandleFunc("POST /identity/token/refresh", h.TokenRefresh)
	mux.HandleFunc("POST /identity/kyc/upgrade",   h.KYCUpgrade_)
	mux.HandleFunc("GET /identity/me",             h.GetMe)
	mux.HandleFunc("PATCH /identity/me/language",  h.PatchLanguage)
	mux.HandleFunc("POST /identity/pin/change",    h.ChangePIN)
	mux.HandleFunc("POST /identity/logout",        h.Logout)
	mux.HandleFunc("GET /identity/users/by-msisdn", h.GetUserByMSISDN) // internal: agent → identity
	mux.HandleFunc("GET /identity/users/{id}",      h.GetUserByID)      // internal: notification → identity
}

