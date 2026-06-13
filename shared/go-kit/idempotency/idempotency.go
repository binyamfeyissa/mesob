package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

const Header = "Idempotency-Key"

func KeyFromRequest(r *http.Request) (string, bool) {
	k := r.Header.Get(Header)
	return k, k != ""
}

func HashResponse(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}
