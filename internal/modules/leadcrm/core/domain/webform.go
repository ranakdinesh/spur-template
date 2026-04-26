package domain

import (
	"time"
)

type LeadForm struct {
	ID        string
	TenantID  string
	Name      string
	Slug      string
	Schema    FormSchema
	Settings  FormSettings
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LeadSubmission struct {
	ID          string
	TenantID    string
	FormID      string
	LeadID      *string
	SubmittedAt time.Time
	Payload     map[string]interface{}
	Headers     SubmissionHeaders
}

type FormSchema struct {
	Fields []FormField `json:"fields"`
}

type FormField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"` // text, email, phone, number
	Required bool   `json:"required"`
}

type FormSettings struct {
	SpamProtection bool   `json:"spam_protection"`
	RedirectURL    string `json:"redirect_url"`
	HoneypotField  string `json:"honeypot_field"`
}

type SubmissionHeaders struct {
	IP        string
	UserAgent string
}
