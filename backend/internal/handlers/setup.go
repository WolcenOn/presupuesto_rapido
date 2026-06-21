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
		if !h.requireDB(w) { return }
		if cfg.BootstrapSecret == "" || r.Header.Get("X-Setup-Token") != cfg.BootstrapSecret { httpx.Error(w, http.StatusForbidden, "setup not allowed"); return }
		var count int
		if err := h.DB.QueryRow(r.Context(), `select count(*) from users where role = 'boss'`).Scan(&count); err != nil { httpx.Error(w, http.StatusInternalServerError, "could not check users"); return }
		if count > 0 { httpx.Error(w, http.StatusConflict, "boss already exists"); return }
		var input struct {
			Name string `json:"name"`
			Email string `json:"email"`
			Secret string `json:"secret"`
			CompanyName string `json:"companyName"`
			CompanyTaxID string `json:"companyTaxId"`
			CompanyEmail string `json:"companyEmail"`
			CompanyPhone string `json:"companyPhone"`
			CompanyAddress string `json:"companyAddress"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil { httpx.Error(w, http.StatusBadRequest, "invalid json"); return }
		input.Name = strings.TrimSpace(input.Name)
		input.Email = strings.ToLower(strings.TrimSpace(input.Email))
		input.CompanyName = strings.TrimSpace(input.CompanyName)
		if input.Name == "" || input.Email == "" || input.Secret == "" || input.CompanyName == "" { httpx.Error(w, http.StatusBadRequest, "name, email, secret and companyName are required"); return }
		hash, err := auth.HashPassword(input.Secret)
		if err != nil { httpx.Error(w, http.StatusBadRequest, err.Error()); return }
		tx, err := h.DB.Begin(r.Context())
		if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not start setup"); return }
		defer tx.Rollback(r.Context())
		var companyID string
		if err = tx.QueryRow(r.Context(), `insert into companies (name, tax_id, email, phone, address) values ($1,$2,$3,$4,$5) returning id::text`, input.CompanyName, input.CompanyTaxID, input.CompanyEmail, input.CompanyPhone, input.CompanyAddress).Scan(&companyID); err != nil { httpx.Error(w, http.StatusInternalServerError, "could not create company"); return }
		var userID string
		if err = tx.QueryRow(r.Context(), `insert into users (company_id, name, email, password_hash, role, is_active) values ($1, $2, $3, $4, 'boss', true) returning id::text`, companyID, input.Name, input.Email, hash).Scan(&userID); err != nil { httpx.Error(w, http.StatusInternalServerError, "could not create user"); return }
		if err = tx.Commit(r.Context()); err != nil { httpx.Error(w, http.StatusInternalServerError, "could not finish setup"); return }
		httpx.JSON(w, http.StatusCreated, map[string]string{"id": userID, "companyId": companyID})
	}
}
