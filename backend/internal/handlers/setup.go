package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"presupuesto-rapido/backend/internal/auth"
	"presupuesto-rapido/backend/internal/httpx"
)

func (h Handler) SetupBoss(cfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.requireDB(w) {
			return
		}
		if cfg.BootstrapSecret == "" || r.Header.Get("X-Setup-Token") != cfg.BootstrapSecret {
			httpx.Error(w, http.StatusForbidden, "setup not allowed")
			return
		}

		var count int
		if err := h.DB.QueryRow(r.Context(), `select count(*) from users where role = 'boss'`).Scan(&count); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not check users")
			return
		}
		if count > 0 {
			httpx.Error(w, http.StatusConflict, "boss already exists")
			return
		}

		var input struct {
			Name   string `json:"name"`
			Email  string `json:"email"`
			Secret string `json:"secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		input.Name = strings.TrimSpace(input.Name)
		input.Email = strings.ToLower(strings.TrimSpace(input.Email))
		if input.Name == "" || input.Email == "" || input.Secret == "" {
			httpx.Error(w, http.StatusBadRequest, "name, email and secret are required")
			return
		}

		hash, err := auth.HashPassword(input.Secret)
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		var id string
		err = h.DB.QueryRow(r.Context(), `insert into users (name, email, password_hash, role, is_active) values ($1, $2, $3, 'boss', true) returning id::text`, input.Name, input.Email, hash).Scan(&id)
		if err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not create user")
			return
		}
		httpx.JSON(w, http.StatusCreated, map[string]string{"id": id})
	}
}
