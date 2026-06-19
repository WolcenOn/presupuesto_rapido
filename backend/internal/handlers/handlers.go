package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"presupuesto-rapido/backend/internal/domain"
	"presupuesto-rapido/backend/internal/httpx"
)

type Handler struct {
	DB        *pgxpool.Pool
	BossEmail string
}

func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{"ok": true, "service": "presupuesto-rapido-api", "time": time.Now().UTC()}
	if h.DB != nil {
		ctx := r.Context()
		if err := h.DB.Ping(ctx); err != nil {
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "database": "unavailable"})
			return
		}
		status["database"] = "ok"
	} else {
		status["database"] = "not_configured"
	}
	httpx.JSON(w, http.StatusOK, status)
}

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := httpx.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h Handler) ListPrices(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) {
		return
	}
	rows, err := h.DB.Query(r.Context(), `select id::text, name, base_price, iva_rate, active, coalesce(updated_by::text, ''), updated_at from price_items where active = true order by name asc`)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list prices")
		return
	}
	defer rows.Close()

	items := []domain.PriceItem{}
	for rows.Next() {
		var it domain.PriceItem
		if err := rows.Scan(&it.ID, &it.Name, &it.BasePrice, &it.IVARate, &it.Active, &it.UpdatedBy, &it.UpdatedAt); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not read prices")
			return
		}
		items = append(items, it)
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h Handler) CreatePrice(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) {
		return
	}
	user, ok := httpx.UserFromContext(r.Context())
	if !ok || user.Role != domain.RoleBoss {
		httpx.Error(w, http.StatusForbidden, "only boss can modify standard prices")
		return
	}
	var input struct {
		Name      string  `json:"name"`
		BasePrice float64 `json:"basePrice"`
		IVARate   float64 `json:"ivaRate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.BasePrice < 0 {
		httpx.Error(w, http.StatusBadRequest, "name and non-negative basePrice are required")
		return
	}
	if input.IVARate == 0 {
		input.IVARate = 21
	}
	var id string
	err := h.DB.QueryRow(r.Context(), `insert into price_items (name, base_price, iva_rate, updated_by) values ($1, $2, $3, $4) returning id::text`, input.Name, input.BasePrice, input.IVARate, user.ID).Scan(&id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not create price item")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) {
		return
	}
	user, ok := httpx.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	query := `select id::text, ref, type, employee_id::text, client_name, client_cif, client_phone, client_address, coalesce(work_order, ''), coalesce(payment_method, ''), base_amount, iva_amount, total_amount, coalesce(pdf_path, ''), sent_to_boss_at, created_at, updated_at from documents`
	args := []any{}
	if user.Role != domain.RoleBoss {
		query += ` where employee_id = $1`
		args = append(args, user.ID)
	}
	query += ` order by created_at desc limit 200`

	rows, err := h.DB.Query(r.Context(), query, args...)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not list documents")
		return
	}
	defer rows.Close()

	docs := []domain.Document{}
	for rows.Next() {
		var d domain.Document
		var sent sql.NullTime
		if err := rows.Scan(&d.ID, &d.Ref, &d.Type, &d.EmployeeID, &d.ClientName, &d.ClientCIF, &d.ClientPhone, &d.ClientAddress, &d.WorkOrder, &d.PaymentMethod, &d.Base, &d.IVA, &d.Total, &d.PDFPath, &sent, &d.CreatedAt, &d.UpdatedAt); err != nil {
			httpx.Error(w, http.StatusInternalServerError, "could not read documents")
			return
		}
		if sent.Valid {
			d.SentToBossAt = &sent.Time
		}
		docs = append(docs, d)
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": docs})
}

func (h Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) {
		return
	}
	user, ok := httpx.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var input struct {
		Ref           string          `json:"ref"`
		Type          string          `json:"type"`
		ClientName    string          `json:"clientName"`
		ClientCIF     string          `json:"clientCif"`
		ClientPhone   string          `json:"clientPhone"`
		ClientAddress string          `json:"clientAddress"`
		WorkOrder     string          `json:"workOrder"`
		PaymentMethod string          `json:"paymentMethod"`
		Base          float64         `json:"base"`
		IVA           float64         `json:"iva"`
		Total         float64         `json:"total"`
		DocumentJSON  json.RawMessage `json:"documentJson"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	input.Ref = strings.TrimSpace(input.Ref)
	input.ClientName = strings.TrimSpace(input.ClientName)
	if input.Ref == "" || input.Type == "" || input.ClientName == "" {
		httpx.Error(w, http.StatusBadRequest, "ref, type and clientName are required")
		return
	}
	if input.Type != string(domain.DocumentBudget) && input.Type != string(domain.DocumentDelivery) && input.Type != string(domain.DocumentInvoice) {
		httpx.Error(w, http.StatusBadRequest, "invalid document type")
		return
	}
	documentJSON := []byte(input.DocumentJSON)
	if len(documentJSON) == 0 {
		documentJSON = []byte(`{}`)
	}
	var id string
	err := h.DB.QueryRow(r.Context(), `
		insert into documents (ref, type, employee_id, client_name, client_cif, client_phone, client_address, work_order, payment_method, base_amount, iva_amount, total_amount, document_json)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		on conflict (employee_id, ref) do update set
			type = excluded.type,
			client_name = excluded.client_name,
			client_cif = excluded.client_cif,
			client_phone = excluded.client_phone,
			client_address = excluded.client_address,
			work_order = excluded.work_order,
			payment_method = excluded.payment_method,
			base_amount = excluded.base_amount,
			iva_amount = excluded.iva_amount,
			total_amount = excluded.total_amount,
			document_json = excluded.document_json,
			updated_at = now()
		returning id::text`, input.Ref, input.Type, user.ID, input.ClientName, input.ClientCIF, input.ClientPhone, input.ClientAddress, input.WorkOrder, input.PaymentMethod, input.Base, input.IVA, input.Total, documentJSON).Scan(&id)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not save document")
		return
	}
	// TODO: enqueue PDF generation and email copy to boss for albaran/factura.
	httpx.JSON(w, http.StatusCreated, map[string]any{"id": id, "synced": true})
}

func (h Handler) requireDB(w http.ResponseWriter) bool {
	if h.DB == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "database not configured")
		return false
	}
	return true
}
