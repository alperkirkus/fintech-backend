package middleware

import (
	"net/http"
	"strings"
)

const maxBodyBytes = 1 << 20

func Validate() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch:
				ct := r.Header.Get("Content-Type")
				if !strings.HasPrefix(ct, "application/json") {
					w.Header().Set("Content-Type", "application/json")
					http.Error(w, `{"error":"Content-Type must be application/json"}`, http.StatusUnsupportedMediaType)
					return
				}
				r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
