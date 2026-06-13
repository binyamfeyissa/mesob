package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(target string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(u), nil
}

func Handler(target string) http.Handler {
	p, err := NewReverseProxy(target)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad gateway", http.StatusBadGateway)
		})
	}
	return p
}
