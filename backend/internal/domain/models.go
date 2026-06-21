package domain

import "time"

type Role string

const (
	RoleOwner    Role = "owner"
	RoleBoss     Role = "boss"
	RoleEmployee Role = "employee"
)

func (r Role) CanManageCompany() bool {
	return r == RoleOwner
}

func (r Role) CanManageUsers() bool {
	return r == RoleOwner || r == RoleBoss
}

func (r Role) CanManagePrices() bool {
	return r == RoleOwner || r == RoleBoss
}

func (r Role) CanCreateDocuments() bool {
	return r == RoleOwner || r == RoleBoss || r == RoleEmployee
}

func (r Role) CanReadCompanyDocuments() bool {
	return r == RoleOwner || r == RoleBoss
}

func (r Role) CanDeleteBusinessData() bool {
	return false
}

type DocumentType string

const (
	DocumentBudget   DocumentType = "presupuesto"
	DocumentDelivery DocumentType = "albaran"
	DocumentInvoice  DocumentType = "factura"
)

type Company struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TaxID     string    `json:"taxId"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	Postal    string    `json:"postal"`
	Province  string    `json:"province"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	ID           string    `json:"id"`
	CompanyID    string    `json:"companyId"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}

type UserInvitation struct {
	ID         string     `json:"id"`
	CompanyID  string     `json:"companyId"`
	Email      string     `json:"email"`
	Role       Role       `json:"role"`
	Token      string     `json:"-"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	AcceptedAt *time.Time `json:"acceptedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type PriceItem struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"companyId"`
	Name      string    `json:"name"`
	BasePrice float64   `json:"basePrice"`
	IVARate   float64   `json:"ivaRate"`
	Active    bool      `json:"active"`
	UpdatedBy string    `json:"updatedBy"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Document struct {
	ID             string       `json:"id"`
	CompanyID      string       `json:"companyId"`
	Ref            string       `json:"ref"`
	Type           DocumentType `json:"type"`
	EmployeeID     string       `json:"employeeId"`
	ClientName     string       `json:"clientName"`
	ClientCIF      string       `json:"clientCif"`
	ClientPhone    string       `json:"clientPhone"`
	ClientAddress  string       `json:"clientAddress"`
	WorkOrder      string       `json:"workOrder"`
	PaymentMethod  string       `json:"paymentMethod"`
	Base           float64      `json:"base"`
	IVA            float64      `json:"iva"`
	Total          float64      `json:"total"`
	DocumentJSON   []byte       `json:"documentJson,omitempty"`
	PDFPath        string       `json:"pdfPath,omitempty"`
	SentToBossAt   *time.Time   `json:"sentToBossAt,omitempty"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type SessionUser struct {
	ID        string `json:"id"`
	CompanyID string `json:"companyId"`
	Email     string `json:"email"`
	Role      Role   `json:"role"`
}
