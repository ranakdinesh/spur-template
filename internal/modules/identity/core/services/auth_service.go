package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/platform/logger"
	"github.com/spurbase/spur/internal/platform/security"
)

type AuthService struct {
	userRepo         ports.UserRepo
	sessionRepo      ports.SessionRepo
	verificationRepo ports.VerificationRepo
	roleRepo         ports.RoleRepo
	permRepo         ports.PermissionRepo
	commPort         ports.CommunicationPort
	log              *logger.Loggerx
}

func NewAuthService(
	userRepo ports.UserRepo,
	sessionRepo ports.SessionRepo,
	verificationRepo ports.VerificationRepo,
	roleRepo ports.RoleRepo,
	permRepo ports.PermissionRepo,
	commPort ports.CommunicationPort,
	log *logger.Loggerx,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		sessionRepo:      sessionRepo,
		verificationRepo: verificationRepo,
		roleRepo:         roleRepo,
		permRepo:         permRepo,
		commPort:         commPort,
		log:              log,
	}
}

// ...

func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*ports.CurrentUser, error) {
	user, err := s.userRepo.GetUser(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	// Fetch Roles
	roles, err := s.roleRepo.ListUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, 0, len(roles))
	permissions := make([]string, 0)

	// Fetch Permissions for each role
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
		perms, err := s.permRepo.ListRolePermissions(ctx, role.ID)
		if err == nil {
			for _, p := range perms {
				permissions = append(permissions, p.FullKey())
			}
		}
	}

	// Deduplicate permissions
	uniquePerms := make(map[string]bool)
	finalPerms := make([]string, 0)
	for _, p := range permissions {
		if !uniquePerms[p] {
			uniquePerms[p] = true
			finalPerms = append(finalPerms, p)
		}
	}

	return &ports.CurrentUser{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		TenantID:     user.TenantID,
		IsSuperAdmin: user.IsSuperAdmin,
		Roles:        roleNames,
		Permissions:  finalPerms,
	}, nil
}

// Login handles Email+Password OR Mobile+Password
func (s *AuthService) Login(ctx context.Context, cmd ports.LoginCmd) (*ports.Session, error) {
	identifier := cmd.Identifier
	if identifier == "" {
		identifier = cmd.Email
	}

	s.log.Info(ctx).Str("identifier", identifier).Msg("Attempting login")

	var user *domain.User
	var err error

	// Determine matching user
	if looksLikeMobile(identifier) {
		user, err = s.userRepo.GetUserByMobile(ctx, identifier)
	} else {
		user, err = s.userRepo.GetUserByEmail(ctx, identifier)
	}

	if err != nil {
		s.log.Warn(ctx).Str("identifier", identifier).Err(err).Msg("Login failed: user not found")
		return nil, errors.New("invalid credentials")
	}

	// Verify Password (Argon2)
	match, err := security.VerifyPassword(cmd.Password, user.PasswordHash)
	if err != nil {
		s.log.Error(ctx).Err(err).Msg("Login failed: password check error")
		return nil, errors.New("internal server error")
	}
	if !match {
		s.log.Warn(ctx).Str("identifier", identifier).Msg("Login failed: invalid password")
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("user is inactive")
	}

	// F4 FIX: Enforce account lock
	if user.IsLocked {
		s.log.Warn(ctx).Str("user_id", user.ID.String()).Msg("Login denied: account is locked")
		return nil, errors.New("account is locked")
	}

	return s.createSession(ctx, user)
}

func (s *AuthService) RequestOTP(ctx context.Context, input ports.RequestOTPInput) error {
	var user *domain.User
	var err error

	if input.Channel == "mobile" || looksLikeMobile(input.Identifier) {
		user, err = s.userRepo.GetUserByMobile(ctx, input.Identifier)
	} else {
		user, err = s.userRepo.GetUserByEmail(ctx, input.Identifier)
	}

	if err != nil {
		// Do not reveal user existence
		s.log.Warn(ctx).Str("identifier", input.Identifier).Msg("RequestOTP: User not found (silent fail)")
		return nil
	}

	// Check for active existing challenge
	kind := domain.VerificationKindEmailOTP
	if input.Channel == "mobile" {
		kind = domain.VerificationKindMobileOTP
	}

	// OTP Generation
	otp, err := generateOTP(6)
	if err != nil {
		return err
	}

	// Hash before storing — store hash, send plain
	otpHash := hashOTP(otp)
	challenge := domain.NewChallenge(user.TenantID, user.ID, kind, otpHash, 2*time.Minute)

	if err := s.verificationRepo.CreateChallenge(ctx, challenge); err != nil {
		return err
	}

	// Send OTP
	return s.commPort.SendOTP(ctx, input.Identifier, input.Channel, otp)
}

func (s *AuthService) ResendOTP(ctx context.Context, input ports.RequestOTPInput) error {
	var user *domain.User
	var err error

	if input.Channel == "mobile" || looksLikeMobile(input.Identifier) {
		user, err = s.userRepo.GetUserByMobile(ctx, input.Identifier)
	} else {
		user, err = s.userRepo.GetUserByEmail(ctx, input.Identifier)
	}

	if err != nil {
		return nil // Silent fail
	}

	kind := domain.VerificationKindEmailOTP
	if input.Channel == "mobile" {
		kind = domain.VerificationKindMobileOTP
	}

	// Check existing valid challenge
	existing, err := s.verificationRepo.GetChallenge(ctx, user.ID, kind)
	if err == nil && existing != nil {
		// Found active challenge.
		// Enforce validity period constraint:
		// "activated only after the validity of the OTP defined"
		// If it's still valid (ExpiresAt > Now), we deny resend.
		if time.Now().Before(existing.ExpiresAt) {
			s.log.Warn(ctx).Str("user_id", user.ID.String()).Msg("ResendOTP: Active OTP exists, denying resend request")
			return errors.New("please wait for current OTP to expire")
		}
	}

	// If no active challenge, or it expired (GetChallenge filters expiration?),
	// GetChallenge query: `expires_at > NOW()`.
	// So if GetChallenge returns nil, it means no active challenge exists.
	// So we can proceed to RequestOTP logic.
	return s.RequestOTP(ctx, input)
}

func (s *AuthService) LoginWithOTP(ctx context.Context, input ports.LoginWithOTPInput) (*ports.Session, error) {
	var user *domain.User
	var err error

	// 1. Identify User
	if looksLikeMobile(input.Identifier) {
		user, err = s.userRepo.GetUserByMobile(ctx, input.Identifier)
	} else {
		user, err = s.userRepo.GetUserByEmail(ctx, input.Identifier)
	}
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 2. Determine Kind (Try both or infer from identifier?)
	// Usually invalid to try both.
	kind := domain.VerificationKindEmailOTP
	if looksLikeMobile(input.Identifier) {
		kind = domain.VerificationKindMobileOTP
	}

	// 3. Verify OTP
	challenge, err := s.verificationRepo.GetChallenge(ctx, user.ID, kind)
	if err != nil || challenge == nil {
		return nil, errors.New("invalid or expired OTP")
	}

	// Hash the input before comparing
	inputHash := hashOTP(input.Code)
	if challenge.TokenHash != inputHash {
		return nil, errors.New("invalid OTP")
	}

	// 4. Mark Consumed
	if err := s.verificationRepo.MarkChallengeConsumed(ctx, challenge.ID); err != nil {
		s.log.Error(ctx).Err(err).Msg("Failed to mark challenge consumed")
		// Proceed anyway? No, security risk (replay).
		// But we already validated.
		// Return error.
		return nil, errors.New("internal server error")
	}

	// 5. Create Session
	return s.createSession(ctx, user)
}

func (s *AuthService) createSession(ctx context.Context, user *domain.User) (*ports.Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}

	session := &ports.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		TenantID:  user.TenantID,
	}

	err = s.sessionRepo.CreateUserSession(ctx, session)
	if err != nil {
		s.log.Error(ctx).Err(err).Msg("Failed to create session")
		return nil, errors.New("internal server error")
	}

	s.log.Info(ctx).Str("user_id", user.ID.String()).Msg("Login successful")
	return session, nil
}

// GetVerificationStatus checks if user is verified or within grace period
func (s *AuthService) GetVerificationStatus(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*ports.VerificationStatus, error) {
	user, err := s.userRepo.GetUser(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	return &ports.VerificationStatus{
		IsVerified:         user.IsVerified(),
		GracePeriodExpired: user.IsGracePeriodExpired(),
	}, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Silent fail
		return nil
	}

	token, _ := generateSessionToken()
	challenge := domain.NewChallenge(user.TenantID, user.ID, domain.VerificationKindPassReset, token, 1*time.Hour)

	if err := s.verificationRepo.CreateChallenge(ctx, challenge); err != nil {
		return err
	}

	// Send reset link
	// link := fmt.Sprintf("https://citual.in/reset-password?token=%s", token)
	return s.commPort.SendOTP(ctx, email, "email", "PASSWORD_RESET_TOKEN:"+token)
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	challenge, err := s.verificationRepo.GetChallengeByToken(ctx, token, domain.VerificationKindPassReset)
	if err != nil || challenge == nil {
		return errors.New("invalid or expired reset token")
	}

	hashedPwd, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, challenge.UserID, challenge.TenantID, hashedPwd); err != nil {
		return err
	}

	// F6 FIX: Invalidate all sessions after password reset
	if err := s.sessionRepo.DeleteUserSessionsByUserID(ctx, challenge.UserID); err != nil {
		// Log but continue — password is already changed
		s.log.Error(ctx).Err(err).Str("user_id", challenge.UserID.String()).
			Msg("Failed to clear sessions after password reset")
	}

	return s.verificationRepo.MarkChallengeConsumed(ctx, challenge.ID)
}

func (s *AuthService) RequestMagicLink(ctx context.Context, identifier string) error {
	var user *domain.User
	var err error
	if looksLikeMobile(identifier) {
		user, err = s.userRepo.GetUserByMobile(ctx, identifier)
	} else {
		user, err = s.userRepo.GetUserByEmail(ctx, identifier)
	}
	if err != nil {
		return nil
	}

	token, _ := generateSessionToken()
	challenge := domain.NewChallenge(user.TenantID, user.ID, domain.VerificationKindMagicLink, token, 15*time.Minute)

	if err := s.verificationRepo.CreateChallenge(ctx, challenge); err != nil {
		return err
	}

	channel := "email"
	if looksLikeMobile(identifier) {
		channel = "mobile"
	}
	return s.commPort.SendOTP(ctx, identifier, channel, "MAGIC_LINK_TOKEN:"+token)
}

func (s *AuthService) LoginWithMagicLink(ctx context.Context, token string) (*ports.Session, error) {
	challenge, err := s.verificationRepo.GetChallengeByToken(ctx, token, domain.VerificationKindMagicLink)
	if err != nil || challenge == nil {
		return nil, errors.New("invalid or expired magic link")
	}

	user, err := s.userRepo.GetUser(ctx, challenge.UserID, challenge.TenantID)
	if err != nil {
		return nil, err
	}

	if err := s.verificationRepo.MarkChallengeConsumed(ctx, challenge.ID); err != nil {
		return nil, err
	}

	return s.createSession(ctx, user)
}

// Existing Methods

func (s *AuthService) GetSession(ctx context.Context, token string) (*ports.Session, error) {
	session, err := s.sessionRepo.GetUserSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.sessionRepo.DeleteUserSession(ctx, token)
		return nil, errors.New("session expired")
	}

	return session, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteUserSession(ctx, token)
}

// Helpers

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateOTP(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		b[i] = digits[num.Int64()]
	}
	return string(b), nil
}

func looksLikeMobile(s string) bool {
	// Simple check: starts with + or contains mostly numbers, no @
	if strings.Contains(s, "@") {
		return false
	}
	match, _ := regexp.MatchString(`^\+?[0-9]{7,15}$`, s)
	return match
}

func hashOTP(otp string) string {
	h := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(h[:])
}
