package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"presupuesto-rapido/backend/internal/auth"
	"presupuesto-rapido/backend/internal/domain"
	"presupuesto-rapido/backend/internal/httpx"
)

func (h Handler) TenantListUsers(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, companyID, ok := h.currentUserAndCompany(r)
	if !ok || !user.Role.CanManageUsers() { httpx.Error(w, http.StatusForbidden, "not allowed"); return }
	rows, err := h.DB.Query(r.Context(), `select id::text, coalesce(company_id::text, ''), name, email, role, is_active, created_at from users where company_id = $1 order by created_at desc`, companyID)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not list users"); return }
	defer rows.Close()
	items := []domain.User{}
	for rows.Next() {
		var u domain.User
		var role string
		if err := rows.Scan(&u.ID, &u.CompanyID, &u.Name, &u.Email, &role, &u.IsActive, &u.CreatedAt); err != nil { httpx.Error(w, http.StatusInternalServerError, "could not read users"); return }
		u.Role = domain.Role(role)
		items = append(items, u)
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h Handler) TenantCreateUser(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	current, companyID, ok := h.currentUserAndCompany(r)
	if !ok || !current.Role.CanManageUsers() { httpx.Error(w, http.StatusForbidden, "not allowed"); return }
	var input struct { Name string `json:"name"`; Email string `json:"email"`; Secret string `json:"secret"`; Role domain.Role `json:"role"`; IsActive *bool `json:"isActive"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { httpx.Error(w, http.StatusBadRequest, "invalid json"); return }
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" || input.Email == "" || input.Secret == "" { httpx.Error(w, http.StatusBadRequest, "name, email and secret are required"); return }
	if input.Role != domain.RoleBoss && input.Role != domain.RoleEmployee { httpx.Error(w, http.StatusBadRequest, "invalid role"); return }
	if current.Role != domain.RoleOwner && input.Role == domain.RoleBoss { httpx.Error(w, http.StatusForbidden, "only owner can create bosses"); return }
	active := true
	if input.IsActive != nil { active = *input.IsActive }
	hash, err := auth.HashPassword(input.Secret)
	if err != nil { httpx.Error(w, http.StatusBadRequest, err.Error()); return }
	var id string
	err = h.DB.QueryRow(r.Context(), `insert into users (company_id, name, email, password_hash, role, is_active) values ($1, $2, $3, $4, $5, $6) returning id::text`, companyID, input.Name, input.Email, hash, input.Role, active).Scan(&id)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not create user"); return }
	httpx.JSON(w, http.StatusCreated, map[string]string{"id": id})
}
