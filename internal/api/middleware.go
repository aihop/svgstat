package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func (a *App) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("Request started")

		next.ServeHTTP(w, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", time.Since(start)).
			Msg("Request completed")
	})
}

func (a *App) getAuthToken(r *http.Request) string {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		return cookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
