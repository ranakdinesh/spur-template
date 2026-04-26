package services_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/services"
	"github.com/spurbase/spur/internal/platform/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAPIKeyRepo
type MockAPIKeyRepo struct {
	mock.Mock
}

func (m *MockAPIKeyRepo) CreateAPIKey(ctx context.Context, key *domain.APIKey) error {
	return m.Called(ctx, key).Error(0)
}
func (m *MockAPIKeyRepo) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*domain.APIKey, error) {
	args := m.Called(ctx, prefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}
func (m *MockAPIKeyRepo) GetAPIKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.APIKey, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}
func (m *MockAPIKeyRepo) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockAPIKeyRepo) ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*domain.APIKey), args.Error(1)
}
func (m *MockAPIKeyRepo) DeleteAPIKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return m.Called(ctx, id, tenantID).Error(0)
}

func TestAPIKeyService_CreateAPIKey(t *testing.T) {
	repo := new(MockAPIKeyRepo)
	log := logger.New(true)
	svc := services.NewAPIKeyService(repo, log)
	ctx := context.Background()
	tenantID := uuid.New()

	repo.On("CreateAPIKey", ctx, mock.AnythingOfType("*domain.APIKey")).Return(nil)

	res, err := svc.CreateAPIKey(ctx, tenantID, "Test Key", []string{"read"})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Secret)
	assert.Contains(t, res.Secret, "ct_")
	assert.Equal(t, "Test Key", res.Key.Name)
	repo.AssertExpectations(t)
}

func TestAPIKeyService_Authenticate_Success(t *testing.T) {
	repo := new(MockAPIKeyRepo)
	log := logger.New(true)
	svc := services.NewAPIKeyService(repo, log)
	ctx := context.Background()

	// 1. Create a real key/hash pair for testing
	// Logic from service: prefix + "." + secret -> sha256 -> hex
	// But since the service does this internally, we can't easily reproduce it without copying logic.
	// OR we can create a key using the service, then mock the repo to return it.
	// Actually, let's just manually construct a valid scenario.
	
	// Let's cheat a bit: We know the logic.
	// fullKey = "ct_prefix.secret"
	// hash = sha256("ct_prefix.secret")
	// We will mock GetAPIKeyByPrefix to return a key with that hash.
	
	// Wait, we need the exact hash logic.
	// Let's use the service to generate a key (mocking create), then use that result for auth.
	
	// Step 1: Mock Create
	repo.On("CreateAPIKey", ctx, mock.AnythingOfType("*domain.APIKey")).Return(nil)
	res, _ := svc.CreateAPIKey(ctx, uuid.New(), "Test", nil)
	
	// Step 2: Setup Mock for Auth
	repo.On("GetAPIKeyByPrefix", ctx, res.Key.Prefix).Return(res.Key, nil)
	repo.On("UpdateAPIKeyLastUsed", mock.Anything, res.Key.ID).Return(nil) // It runs in background, so use Any Context

	// Act
	key, err := svc.Authenticate(ctx, res.Secret)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, res.Key.ID, key.ID)
}

func TestAPIKeyService_Authenticate_Invalid(t *testing.T) {
	repo := new(MockAPIKeyRepo)
	log := logger.New(true)
	svc := services.NewAPIKeyService(repo, log)
	ctx := context.Background()

	// Case: Invalid Format
	_, err := svc.Authenticate(ctx, "invalidformat")
	assert.Error(t, err)

	// Case: Key Not Found
	repo.On("GetAPIKeyByPrefix", ctx, "unknown").Return(nil, nil)
	_, err = svc.Authenticate(ctx, "unknown.secret")
	assert.Error(t, err)
}

func TestAPIKeyService_Authenticate_Expired(t *testing.T) {
	repo := new(MockAPIKeyRepo)
	log := logger.New(true)
	svc := services.NewAPIKeyService(repo, log)
	ctx := context.Background()

	// Calculate valid hash for the test key
	hash := sha256.Sum256([]byte("ct_test.secret"))
	validHash := hex.EncodeToString(hash[:])

	expiredKey := &domain.APIKey{
		ID:        uuid.New(),
		Prefix:    "ct_test",
		KeyHash:   validHash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}

	repo.On("GetAPIKeyByPrefix", ctx, "ct_test").Return(expiredKey, nil)

	_, err := svc.Authenticate(ctx, "ct_test.secret")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}
