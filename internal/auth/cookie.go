package auth

import (
	"net/http"
	"time"
)

const (
	RefreshCookieName = "rb_refresh"
	// Path "/" so the cookie is included in document requests, enabling SSR
	// middleware to forward it when calling /api/auth/refresh server-side.
	RefreshCookiePath = "/"
)

func SetRefreshCookie(w http.ResponseWriter, token string, devMode bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     RefreshCookiePath,
		MaxAge:   int(RefreshTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   !devMode,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearRefreshCookie(w http.ResponseWriter, devMode bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     RefreshCookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   !devMode,
		SameSite: http.SameSiteLaxMode,
	})
}

func ReadRefreshCookie(r *http.Request) string {
	c, err := r.Cookie(RefreshCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
