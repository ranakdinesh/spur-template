package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidFirstName = errors.New("invalid first name")
	ErrInvalidLastName  = errors.New("invalid last name")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidMobile    = errors.New("invalid mobile")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUserInactive     = errors.New("user is inactive")
)

type User struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Mobile       string    `json:"mobile"` // Added
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsSuperAdmin bool      `json:"is_super_admin"`
	AuthzVersion int       `json:"authz_version"`
	IsActive     bool      `json:"is_active"`
	IsLocked     bool      `json:"is_locked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Verification Status
	EmailVerifiedAt            *time.Time `json:"email_verified_at,omitempty"`
	MobileVerifiedAt           *time.Time `json:"mobile_verified_at,omitempty"`
	VerifiedAt                 *time.Time `json:"verified_at,omitempty"`
	VerificationGracePeriodEnd *time.Time `json:"verification_grace_period_end,omitempty"`
}

// NewUser Factory
func NewUser(tenantID uuid.UUID, firstName, lastName, email, mobile string) (*User, error) {
	if strings.TrimSpace(firstName) == "" {
		return nil, ErrInvalidFirstName
	}
	if strings.TrimSpace(lastName) == "" {
		return nil, ErrInvalidLastName
	}
	if !looksLikeEmail(email) {
		return nil, ErrInvalidEmail
	}

	return &User{
		ID:        uuid.New(),
		TenantID:  tenantID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     strings.ToLower(email),
		Mobile:    mobile,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

// Rich Domain Methods (Keep logic inside Domain!)

func (u *User) UpdateName(firstName, lastName string) error {
	if strings.TrimSpace(firstName) == "" {
		return ErrInvalidFirstName
	}
	u.FirstName = firstName
	u.LastName = lastName
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) IsVerified() bool {
	return u.VerifiedAt != nil || u.EmailVerifiedAt != nil || u.MobileVerifiedAt != nil
}

func (u *User) IsGracePeriodExpired() bool {
	if u.IsVerified() {
		return false
	}
	if u.VerificationGracePeriodEnd == nil {
		return false // Or true? If no end set, maybe we haven't started enforcement.
	}
	return time.Now().UTC().After(*u.VerificationGracePeriodEnd)
}

// Validation Helpers
func looksLikeEmail(s string) bool {
	return strings.Contains(s, "@") && strings.Contains(s, ".")
}
