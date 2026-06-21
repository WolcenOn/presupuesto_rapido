package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"presupuesto-rapido/backend/internal/domain"
	"presupuesto-rapido/backend/internal/httpx"
)

func (h Handler) TenantListPrices(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	_, companyID, ok := h.currentUserAndCompany(r)
	if !ok { httpx.Error(w, http.StatusUnauthorized, "authentication required"); return }
	rows, err := h.DB.Query(r.Context(), `select id::text, coalesce(company_id::text, ''), name, base_price, iva_rate, active, coalesce(updated_by::text, ''), updated_at from price_items where company_id = $1 and active = true order by name asc`, companyID)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not list prices"); return }
	defer rows.Close()
	items := []domain.PriceItem{}
	for rows.Next() {
		var it domain.PriceItem
		if err := rows.Scan(&it.ID, &it.CompanyID, &it.Name, &it.BasePrice, &it.IVARate, &it.Active, &it.UpdatedBy, &it.UpdatedAt); err != nil { httpx.Error(w, http.StatusInternalServerError, "could not read prices"); return }
		items = append(items, it)
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h Handler) TenantGetPrice(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	_, companyID, ok := h.currentUserAndCompany(r)
	if !ok { httpx.Error(w, http.StatusUnauthorized, "authentication required"); return }
	id := strings.TrimSpace(r.PathValue("id"))
	var it domain.PriceItem
	err := h.DB.QueryRow(r.Context(), `select id::text, coalesce(company_id::text, ''), name, base_price, iva_rate, active, coalesce(updated_by::text, ''), updated_at from price_items where id = $1 and company_id = $2`, id, companyID).Scan(&it.ID, &it.CompanyID, &it.Name, &it.BasePrice, &it.IVARate, &it.Active, &it.UpdatedBy, &it.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) { httpx.Error(w, http.StatusNotFound, "price item not found"); return }
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not read price item"); return }
	httpx.JSON(w, http.StatusOK, it)
}

func (h Handler) TenantCreatePrice(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, companyID, ok := h.currentUserAndCompany(r)
	if !ok || !user.Role.CanManagePrices() { httpx.Error(w, http.StatusForbidden, "not allowed"); return }
	var input struct { Name string `json:"name"`; BasePrice float64 `json:"basePrice"`; IVARate float64 `json:"ivaRate"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { httpx.Error(w, http.StatusBadRequest, "invalid json"); return }
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.BasePrice < 0 { httpx.Error(w, http.StatusBadRequest, "name and non-negative basePrice are required"); return }
	if input.IVARate == 0 { input.IVARate = 21 }
	var id string
	err := h.DB.QueryRow(r.Context(), `insert into price_items (company_id, name, base_price, iva_rate, updated_by) values ($1, $2, $3, $4, $5) returning id::text`, companyID, input.Name, input.BasePrice, input.IVARate, user.ID).Scan(&id)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not create price item"); return }
	httpx.JSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h Handler) TenantUpdatePrice(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, companyID, ok := h.currentUserAndCompany(r)
	if !ok || !user.Role.CanManagePrices() { httpx.Error(w, http.StatusForbidden, "not allowed"); return }
	id := strings.TrimSpace(r.PathValue("id"))
	var input struct { Name string `json:"name"`; BasePrice float64 `json:"basePrice"`; IVARate float64 `json:"ivaRate"`; Active *bool `json:"active"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { httpx.Error(w, http.StatusBadRequest, "invalid json"); return }
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.BasePrice < 0 || input.IVARate < 0 { httpx.Error(w, http.StatusBadRequest, "invalid price data"); return }
	active := true
	if input.Active != nil { active = *input.Active }
	cmd, err := h.DB.Exec(r.Context(), `update price_items set name = $1, base_price = $2, iva_rate = $3, active = $4, updated_by = $5, updated_at = now() where id = $6 and company_id = $7`, input.Name, input.BasePrice, input.IVARate, active, user.ID, id, companyID)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not update price item"); return }
	if cmd.RowsAffected() == 0 { httpx.Error(w, http.StatusNotFound, "price item not found"); return }
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}
