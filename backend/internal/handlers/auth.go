package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"presupuesto-rapido/backend/internal/auth"
	"presupuesto-rapido/backend/internal/domain"
	"presupuesto-rapido/backend/internal/httpx"
)

type AuthConfig struct {
	JWTSecret       string
	BootstrapSecret string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CookieSecure    bool
}

func (h Handler) Login(cfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.requireDB(w) {
			return
		}
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		input.Email = strings.ToLower(strings.TrimSpace(input.Email))
		if input.Email == "" || input.Password == "" {
			httpx.Error(w, http.StatusBadRequest, "email and password are required")
			return
		}

		var user domain.User
		var role string
		err := h.DB.QueryRow(r.Context(), `select id::text, name, email, password_hash, role, is_active, created_at from users where lower(email) = $1`, input.Email).
			Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &role, &user.IsActive, &user.CreatedAt)
		user.Role = domain.Role(role)
		if err != nil || !user.IsActive || !auth.VerifyPassword(input.Password, user.PasswordHash) {
			httpx.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		sessionUser := domain.SessionUser{ID: user.ID, Email: user.Email, Role: user.Role}
		accessToken, err := auth.CreateAccessToken(sessionUser, cfg.JWTSecret, cfg.AccessTokenTTL)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not create access token")
			return
		}
		refreshPlain, refreshHash, err := auth.NewRefreshToken()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not create refresh token")
			return
		}
		expiresAt := time.Now().UTC().Add(cfg.RefreshTokenTTL)
		_, err = h.DB.Exec(r.Context(), `insert into refresh_tokens (user_id, token_hash, expires_at) values ($1, $2, $3)`, user.ID, refreshHash, expiresAt)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not persist refresh token")
			return
		}
		setRefreshCookie(w, refreshPlain, expiresAt, cfg.CookieSecure)
		httpx.JSON(w, http.StatusOK, map[string]any{"accessToken": accessToken, "user": sessionUser})
	}
}

func (h Handler) Refresh(cfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.requireDB(w) {
			return
		}
		cookie, err := r.Cookie("amp_refresh_token")
		if err != nil || cookie.Value == "" {
			httpx.Error(w, http.StatusUnauthorized, "refresh token required")
			return
		}
		hash := auth.HashRefreshToken(cookie.Value)
		var user domain.SessionUser
		var tokenID string
		var role string
		err = h.DB.QueryRow(r.Context(), `
			select rt.id::text, u.id::text, u.email, u.role
			from refresh_tokens rt
			join users u on u.id = rt.user_id
			where rt.token_hash = $1 and rt.revoked_at is null and rt.expires_at > now() and u.is_active = true`, hash).
			Scan(&tokenID, &user.ID, &user.Email, &role)
		user.Role = domain.Role(role)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}

		newPlain, newHash, err := auth.NewRefreshToken()
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not rotate refresh token")
			return
		}
		expiresAt := time.Now().UTC().Add(cfg.RefreshTokenTTL)
		tx, err := h.DB.Begin(r.Context())
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not refresh session")
			return
		}
		defer tx.Rollback(r.Context())
		if _, err := tx.Exec(r.Context(), `update refresh_tokens set revoked_at = now() where id = $1`, tokenID); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not revoke old token")
			return
		}
		if _, err := tx.Exec(r.Context(), `insert into refresh_tokens (user_id, token_hash, expires_at) values ($1, $2, $3)`, user.ID, newHash, expiresAt); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not persist new token")
			return
		}
		if err := tx.Commit(r.Context()); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not commit refresh")
			return
		}
		accessToken, err := auth.CreateAccessToken(user, cfg.JWTSecret, cfg.AccessTokenTTL)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not create access token")
			return
		}
		setRefreshCookie(w, newPlain, expiresAt, cfg.CookieSecure)
		httpx.JSON(w, http.StatusOK, map[string]any{"accessToken": accessToken, "user": user})
	}
}

func (h Handler) Logout(cfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.DB != nil {
			if cookie, err := r.Cookie("amp_refresh_token"); err == nil && cookie.Value != "" {
				hash := auth.HashRefreshToken(cookie.Value)
				_, _ = h.DB.Exec(r.Context(), `update refresh_tokens set revoked_at = now() where token_hash = $1 and revoked_at is null`, hash)
			}
		}
		clearRefreshCookie(w, cfg.CookieSecure)
		httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func setRefreshCookie(w http.ResponseWriter, value string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "amp_refresh_token",
		Value:    value,
		Path:     "/api/auth",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearRefreshCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "amp_refresh_token",
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
