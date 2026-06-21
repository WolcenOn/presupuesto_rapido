package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"presupuesto-rapido/backend/internal/domain"
	"presupuesto-rapido/backend/internal/httpx"
)

func (h Handler) GetCompany(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, ok := httpx.UserFromContext(r.Context())
	if !ok || user.CompanyID == "" { httpx.Error(w, http.StatusUnauthorized, "authentication required"); return }
	var c domain.Company
	err := h.DB.QueryRow(r.Context(), `select id::text, name, coalesce(tax_id,''), coalesce(email,''), coalesce(phone,''), coalesce(address,''), coalesce(city,''), coalesce(postal,''), coalesce(province,''), country, created_at, updated_at from companies where id = $1`, user.CompanyID).Scan(&c.ID, &c.Name, &c.TaxID, &c.Email, &c.Phone, &c.Address, &c.City, &c.Postal, &c.Province, &c.Country, &c.CreatedAt, &c.UpdatedAt)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not read company"); return }
	httpx.JSON(w, http.StatusOK, c)
}

func (h Handler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, ok := httpx.UserFromContext(r.Context())
	if !ok || user.CompanyID == "" { httpx.Error(w, http.StatusUnauthorized, "authentication required"); return }
	if user.Role != domain.RoleBoss { httpx.Error(w, http.StatusForbidden, "only boss can update company"); return }
	var input domain.Company
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { httpx.Error(w, http.StatusBadRequest, "invalid json"); return }
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" { httpx.Error(w, http.StatusBadRequest, "company name is required"); return }
	_, err := h.DB.Exec(r.Context(), `update companies set name=$1, tax_id=$2, email=$3, phone=$4, address=$5, city=$6, postal=$7, province=$8, country=$9, updated_at=now() where id=$10`, input.Name, input.TaxID, input.Email, input.Phone, input.Address, input.City, input.Postal, input.Province, valueOr(input.Country, "ES"), user.CompanyID)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not update company"); return }
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func valueOr(v, fallback string) string {
	if strings.TrimSpace(v) == "" { return fallback }
	return v
}
