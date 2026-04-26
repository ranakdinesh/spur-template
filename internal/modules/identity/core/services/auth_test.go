package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/modules/identity/core/services"
	"github.com/spurbase/spur/internal/platform/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) GetUser(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id, tenantID)
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) GetUserByMobile(ctx context.Context, mobile string) (*domain.User, error) {
	args := m.Called(ctx, mobile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) ListUsers(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*domain.User), args.Error(1)
}
func (m *MockUserRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepo) DeleteUser(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return m.Called(ctx, id, tenantID).Error(0)
}
func (m *MockUserRepo) DeleteUserByEmail(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *MockUserRepo) DeleteUserByMobile(ctx context.Context, mobile string) error {
	return m.Called(ctx, mobile).Error(0)
}
func (m *MockUserRepo) UpdateUserLockStatus(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, locked bool) error {
	return m.Called(ctx, userID, tenantID, locked).Error(0)
}

type MockVerificationRepo struct {
	mock.Mock
}

func (m *MockVerificationRepo) CreateChallenge(ctx context.Context, challenge *domain.VerificationChallenge) error {
	return m.Called(ctx, challenge).Error(0)
}
func (m *MockVerificationRepo) GetChallenge(ctx context.Context, userID uuid.UUID, kind domain.VerificationKind) (*domain.VerificationChallenge, error) {
	args := m.Called(ctx, userID, kind)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VerificationChallenge), args.Error(1)
}
func (m *MockVerificationRepo) GetChallengeByToken(ctx context.Context, token string, kind domain.VerificationKind) (*domain.VerificationChallenge, error) {
	args := m.Called(ctx, token, kind)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.VerificationChallenge), args.Error(1)
}
func (m *MockVerificationRepo) MarkChallengeConsumed(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockVerificationRepo) DeleteExpiredChallenges(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, hash string) error {
	return m.Called(ctx, userID, tenantID, hash).Error(0)
}

type MockCommPort struct {
	mock.Mock
}

func (m *MockCommPort) SendOTP(ctx context.Context, recipient, channel, code string) error {
	return m.Called(ctx, recipient, channel, code).Error(0)
}

type MockSessionRepo struct {
	mock.Mock
}

func (m *MockSessionRepo) CreateUserSession(ctx context.Context, s *ports.Session) error {
	return m.Called(ctx, s).Error(0)
}
func (m *MockSessionRepo) GetUserSessionByToken(ctx context.Context, token string) (*ports.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.Session), args.Error(1)
}
func (m *MockSessionRepo) DeleteUserSession(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *MockSessionRepo) DeleteUserSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

// ... Existing Tests ...

func TestAuthService_RequestPasswordReset(t *testing.T) {
	userRepo := new(MockUserRepo)
	verRepo := new(MockVerificationRepo)
	commPort := new(MockCommPort)
	log := logger.New(true)
	svc := services.NewAuthService(userRepo, nil, verRepo, nil, nil, commPort, log)
	ctx := context.Background()

	user := &domain.User{ID: uuid.New(), TenantID: uuid.New(), Email: "reset@example.com"}

	userRepo.On("GetUserByEmail", ctx, "reset@example.com").Return(user, nil)
	verRepo.On("CreateChallenge", ctx, mock.Anything).Return(nil)
	commPort.On("SendOTP", ctx, "reset@example.com", "email", mock.Anything).Return(nil)

	err := svc.RequestPasswordReset(ctx, "reset@example.com")
	assert.NoError(t, err)
}

func TestAuthService_ResetPassword(t *testing.T) {
	userRepo := new(MockUserRepo)
	verRepo := new(MockVerificationRepo)
	log := logger.New(true)
	svc := services.NewAuthService(userRepo, nil, verRepo, nil, nil, nil, log)
	ctx := context.Background()

	challenge := &domain.VerificationChallenge{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		TenantID: uuid.New(),
		Kind:     domain.VerificationKindPassReset,
	}

	verRepo.On("GetChallengeByToken", ctx, "valid_token", domain.VerificationKindPassReset).Return(challenge, nil)
	userRepo.On("UpdatePassword", ctx, challenge.UserID, challenge.TenantID, mock.Anything).Return(nil)
	verRepo.On("MarkChallengeConsumed", ctx, challenge.ID).Return(nil)

	err := svc.ResetPassword(ctx, "valid_token", "newSecret123!")
	assert.NoError(t, err)
}

func TestAuthService_GetVerificationStatus(t *testing.T) {
	userRepo := new(MockUserRepo)
	log := logger.New(true)
	svc := services.NewAuthService(userRepo, nil, nil, nil, nil, nil, log)
	ctx := context.Background()

	uid := uuid.New()
	tid := uuid.New()
	// Case 1: Verified
	user := &domain.User{ID: uid, VerifiedAt: func() *time.Time { t := time.Now(); return &t }()}
	userRepo.On("GetUser", ctx, uid, tid).Return(user, nil).Once()

	status, err := svc.GetVerificationStatus(ctx, uid, tid)
	assert.NoError(t, err)
	assert.True(t, status.IsVerified)

	// Case 2: Not Verified, Grace Period Active
	user2 := &domain.User{ID: uid, VerificationGracePeriodEnd: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }()}
	userRepo.On("GetUser", ctx, uid, tid).Return(user2, nil).Once()

	status2, _ := svc.GetVerificationStatus(ctx, uid, tid)
	assert.False(t, status2.IsVerified)
	assert.False(t, status2.GracePeriodExpired)
}


func TestAuthService_RequestOTP(t *testing.T) {
	userRepo := new(MockUserRepo)
	verRepo := new(MockVerificationRepo)
	commPort := new(MockCommPort)
	sessRepo := new(MockSessionRepo)
	log := logger.New(true) // Assuming logger constructor exists or use nil context log

	svc := services.NewAuthService(userRepo, sessRepo, verRepo, nil, nil, commPort, log)

	ctx := context.Background()
	user := &domain.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Email:    "test@example.com",
		IsActive: true,
	}

	// Setup expectations
	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)
	verRepo.On("CreateChallenge", ctx, mock.AnythingOfType("*domain.VerificationChallenge")).Return(nil)
	commPort.On("SendOTP", ctx, "test@example.com", "email", mock.AnythingOfType("string")).Return(nil)

	err := svc.RequestOTP(ctx, ports.RequestOTPInput{
		Identifier: "test@example.com",
		Channel:    "email",
	})

	assert.NoError(t, err)
	verRepo.AssertExpectations(t)
	commPort.AssertExpectations(t)
}

func TestAuthService_ResendOTP_DenyIfValid(t *testing.T) {
	userRepo := new(MockUserRepo)
	verRepo := new(MockVerificationRepo)
	commPort := new(MockCommPort)
	sessRepo := new(MockSessionRepo)
	log := logger.New(true)

	svc := services.NewAuthService(userRepo, sessRepo, verRepo, nil, nil, commPort, log)
	ctx := context.Background()
	user := &domain.User{ID: uuid.New(), Email: "test@example.com"}

	validChallenge := &domain.VerificationChallenge{
		ID:        uuid.New(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(1 * time.Minute), // Valid
	}

	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)
	verRepo.On("GetChallenge", ctx, user.ID, domain.VerificationKindEmailOTP).Return(validChallenge, nil)

	err := svc.ResendOTP(ctx, ports.RequestOTPInput{Identifier: "test@example.com", Channel: "email"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "please wait")
}
