package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"presupuesto-rapido/backend/internal/auth"
	"presupuesto-rapido/backend/internal/domain"
	"presupuesto-rapido/backend/internal/httpx"
)

func (h Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) {
		return
	}
	rows, err := h.DB.Query(r.Context(), `select id::text, name, email, role, is_active, created_at from users order by created_at desc`)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list users")
		return
	}
	defer rows.Close()

	items := []domain.User{}
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not read users")
			return
		}
		items = append(items, u)
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) {
		return
	}
	var input struct {
		Name     string      `json:"name"`
		Email    string      `json:"email"`
		Secret   string      `json:"secret"`
		Role     domain.Role `json:"role"`
		IsActive *bool       `json:"isActive"`
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
	if input.Role != domain.RoleBoss && input.Role != domain.RoleEmployee {
		httpx.Error(w, http.StatusBadRequest, "invalid role")
		return
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	hash, err := auth.HashPassword(input.Secret)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	var id string
	err = h.DB.QueryRow(r.Context(), `insert into users (name, email, password_hash, role, is_active) values ($1, $2, $3, $4, $5) returning id::text`, input.Name, input.Email, hash, input.Role, active).Scan(&id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create user")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]string{"id": id})
}
