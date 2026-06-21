package handlers

import (
	"errors"
	"net/http"

	"presupuesto-rapido/backend/internal/domain"
	"presupuesto-rapido/backend/internal/httpx"
)

var errCompanyRequired = errors.New("company required")

func (h Handler) currentUserAndCompany(r *http.Request) (domain.SessionUser, string, bool) {
	user, ok := httpx.UserFromContext(r.Context())
	if !ok {
		return domain.SessionUser{}, "", false
	}
	if user.CompanyID != "" {
		return user, user.CompanyID, true
	}
	if h.DB == nil || user.ID == "" {
		return user, "", false
	}
	var companyID string
	if err := h.DB.QueryRow(r.Context(), `select coalesce(company_id::text, '') from users where id = $1 and is_active = true`, user.ID).Scan(&companyID); err != nil || companyID == "" {
		return user, "", false
	}
	user.CompanyID = companyID
	return user, companyID, true
}
