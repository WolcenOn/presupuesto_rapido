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

func (h Handler) TenantListDocuments(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, companyID, ok := h.currentUserAndCompany(r)
	if !ok { httpx.Error(w, http.StatusUnauthorized, "authentication required"); return }
	query := `select id::text, coalesce(company_id::text, ''), ref, type, employee_id::text, client_name, client_cif, client_phone, client_address, coalesce(work_order, ''), coalesce(payment_method, ''), base_amount, iva_amount, total_amount, coalesce(pdf_path, ''), sent_to_boss_at, created_at, updated_at from documents where company_id = $1`
	args := []any{companyID}
	if !user.Role.CanReadCompanyDocuments() {
		query += ` and employee_id = $2`
		args = append(args, user.ID)
	}
	query += ` order by created_at desc limit 200`
	rows, err := h.DB.Query(r.Context(), query, args...)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not list documents"); return }
	defer rows.Close()
	docs := []domain.Document{}
	for rows.Next() {
		var d domain.Document
		var docType string
		var sent sql.NullTime
		if err := rows.Scan(&d.ID, &d.CompanyID, &d.Ref, &docType, &d.EmployeeID, &d.ClientName, &d.ClientCIF, &d.ClientPhone, &d.ClientAddress, &d.WorkOrder, &d.PaymentMethod, &d.Base, &d.IVA, &d.Total, &d.PDFPath, &sent, &d.CreatedAt, &d.UpdatedAt); err != nil { httpx.Error(w, http.StatusInternalServerError, "could not read documents"); return }
		d.Type = domain.DocumentType(docType)
		if sent.Valid { d.SentToBossAt = &sent.Time }
		docs = append(docs, d)
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": docs})
}

func (h Handler) TenantGetDocument(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, companyID, ok := h.currentUserAndCompany(r)
	if !ok { httpx.Error(w, http.StatusUnauthorized, "authentication required"); return }
	id := strings.TrimSpace(r.PathValue("id"))
	doc, err := h.loadTenantDocument(r, id, companyID)
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) { httpx.Error(w, http.StatusNotFound, "document not found"); return }
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not read document"); return }
	if !user.Role.CanReadCompanyDocuments() && doc.EmployeeID != user.ID { httpx.Error(w, http.StatusForbidden, "document does not belong to this user"); return }
	httpx.JSON(w, http.StatusOK, doc)
}

func (h Handler) TenantCreateDocument(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, companyID, ok := h.currentUserAndCompany(r)
	if !ok || !user.Role.CanCreateDocuments() { httpx.Error(w, http.StatusForbidden, "not allowed"); return }
	var input struct { Ref string `json:"ref"`; Type string `json:"type"`; ClientName string `json:"clientName"`; ClientCIF string `json:"clientCif"`; ClientPhone string `json:"clientPhone"`; ClientAddress string `json:"clientAddress"`; WorkOrder string `json:"workOrder"`; PaymentMethod string `json:"paymentMethod"`; Base float64 `json:"base"`; IVA float64 `json:"iva"`; Total float64 `json:"total"`; DocumentJSON json.RawMessage `json:"documentJson"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { httpx.Error(w, http.StatusBadRequest, "invalid json"); return }
	input.Ref = strings.TrimSpace(input.Ref)
	input.ClientName = strings.TrimSpace(input.ClientName)
	if input.Ref == "" || input.Type == "" || input.ClientName == "" { httpx.Error(w, http.StatusBadRequest, "ref, type and clientName are required"); return }
	if input.Type != string(domain.DocumentBudget) && input.Type != string(domain.DocumentDelivery) && input.Type != string(domain.DocumentInvoice) { httpx.Error(w, http.StatusBadRequest, "invalid document type"); return }
	documentJSON := []byte(input.DocumentJSON)
	if len(documentJSON) == 0 { documentJSON = []byte(`{}`) }
	var id string
	err := h.DB.QueryRow(r.Context(), `insert into documents (company_id, ref, type, employee_id, client_name, client_cif, client_phone, client_address, work_order, payment_method, base_amount, iva_amount, total_amount, document_json) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) on conflict (employee_id, ref) do update set type = excluded.type, client_name = excluded.client_name, client_cif = excluded.client_cif, client_phone = excluded.client_phone, client_address = excluded.client_address, work_order = excluded.work_order, payment_method = excluded.payment_method, base_amount = excluded.base_amount, iva_amount = excluded.iva_amount, total_amount = excluded.total_amount, document_json = excluded.document_json, updated_at = now() returning id::text`, companyID, input.Ref, input.Type, user.ID, input.ClientName, input.ClientCIF, input.ClientPhone, input.ClientAddress, input.WorkOrder, input.PaymentMethod, input.Base, input.IVA, input.Total, documentJSON).Scan(&id)
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not save document"); return }
	if input.Type == string(domain.DocumentDelivery) || input.Type == string(domain.DocumentInvoice) { _ = h.queueBossEmail(r, id) }
	httpx.JSON(w, http.StatusCreated, map[string]any{"id": id, "synced": true})
}

func (h Handler) TenantQueueDocumentForBoss(w http.ResponseWriter, r *http.Request) {
	if !h.requireDB(w) { return }
	user, companyID, ok := h.currentUserAndCompany(r)
	if !ok { httpx.Error(w, http.StatusUnauthorized, "authentication required"); return }
	id := strings.TrimSpace(r.PathValue("id"))
	doc, err := h.loadTenantDocument(r, id, companyID)
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) { httpx.Error(w, http.StatusNotFound, "document not found"); return }
	if err != nil { httpx.Error(w, http.StatusInternalServerError, "could not read document"); return }
	if !user.Role.CanReadCompanyDocuments() && doc.EmployeeID != user.ID { httpx.Error(w, http.StatusForbidden, "document does not belong to this user"); return }
	if doc.Type != domain.DocumentDelivery && doc.Type != domain.DocumentInvoice { httpx.Error(w, http.StatusBadRequest, "only albaran and factura can be queued for boss email"); return }
	if err := h.queueBossEmail(r, id); err != nil { httpx.Error(w, http.StatusInternalServerError, "could not queue boss email"); return }
	httpx.JSON(w, http.StatusAccepted, map[string]bool{"queued": true})
}

func (h Handler) loadTenantDocument(r *http.Request, id string, companyID string) (domain.Document, error) {
	var d domain.Document
	var docType string
	var sent sql.NullTime
	err := h.DB.QueryRow(r.Context(), `select id::text, coalesce(company_id::text, ''), ref, type, employee_id::text, client_name, client_cif, client_phone, client_address, coalesce(work_order, ''), coalesce(payment_method, ''), base_amount, iva_amount, total_amount, document_json, coalesce(pdf_path, ''), sent_to_boss_at, created_at, updated_at from documents where id = $1 and company_id = $2`, id, companyID).Scan(&d.ID, &d.CompanyID, &d.Ref, &docType, &d.EmployeeID, &d.ClientName, &d.ClientCIF, &d.ClientPhone, &d.ClientAddress, &d.WorkOrder, &d.PaymentMethod, &d.Base, &d.IVA, &d.Total, &d.DocumentJSON, &d.PDFPath, &sent, &d.CreatedAt, &d.UpdatedAt)
	d.Type = domain.DocumentType(docType)
	if sent.Valid { d.SentToBossAt = &sent.Time }
	return d, err
}
