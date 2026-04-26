package domain

type LeadIdentityType string

const (
	IdentityTypeEmail      LeadIdentityType = "email"
	IdentityTypePhone      LeadIdentityType = "phone"
	IdentityTypeMetaPSID   LeadIdentityType = "meta_psid"
	IdentityTypeLinkedInID LeadIdentityType = "linkedin_id"
	IdentityTypeCookieID   LeadIdentityType = "cookie_id"
)

type LeadIdentity struct {
	ID       string
	TenantID string
	LeadID   string
	Type     LeadIdentityType
	Value    string
}
