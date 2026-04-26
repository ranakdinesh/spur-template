package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spurbase/spur/internal/modules/identity/core/domain"
	"github.com/spurbase/spur/internal/modules/identity/core/ports"
	"github.com/spurbase/spur/internal/platform/logger"
)

type APIKeyService struct {
	repo ports.APIKeyRepo
	log  *logger.Loggerx
}

func NewAPIKeyService(repo ports.APIKeyRepo, log *logger.Loggerx) *APIKeyService {
	return &APIKeyService{
		repo: repo,
		log:  log,
	}
}

type CreateAPIKeyInput struct {
	TenantID       uuid.UUID
	Name           string
	Type           string // secret, publishable
	Scopes         []string
	AllowedOrigins []string
}

func (s *APIKeyService) CreateAPIKey(ctx context.Context, input ports.CreateAPIKeyCmd) (*ports.CreateAPIKeyResult, error) {
	// 1. Generate Prefix and Secret
	// Publishable keys get a different prefix for easy identification
	prefixRoot := "ct_"
	if input.Type == "publishable" {
		prefixRoot = "pk_"
	}
	prefixSuffix, err := generateRandomString(8)
	if err != nil {
		return nil, fmt.Errorf("generate key prefix: %w", err)
	}
	secret, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("generate key secret: %w", err)
	}
	prefix := prefixRoot + prefixSuffix
	fullKey := prefix + "." + secret

	// 2. Hash Secret
	hash := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hash[:])

	// 3. Store
	apiKey := &domain.APIKey{
		ID:             uuid.New(),
		TenantID:       input.TenantID,
		Name:           input.Name,
		Type:           input.Type,
		Prefix:         prefix,
		KeyHash:        keyHash,
		Scopes:         input.Scopes,
		AllowedOrigins: input.AllowedOrigins,
		ExpiresAt:      time.Now().Add(365 * 24 * time.Hour), // 1 year default
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.repo.CreateAPIKey(ctx, apiKey); err != nil {
		return nil, err
	}

	return &ports.CreateAPIKeyResult{
		Key:    apiKey,
		Secret: fullKey,
	}, nil
}

func (s *APIKeyService) Authenticate(ctx context.Context, fullKey string, currentOrigin string) (*domain.APIKey, error) {
	parts := strings.Split(fullKey, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid api key format")
	}
	prefix := parts[0]

	// 1. Find key by prefix
	apiKey, err := s.repo.GetAPIKeyByPrefix(ctx, prefix)
	if err != nil || apiKey == nil {
		return nil, fmt.Errorf("invalid api key")
	}

	// 2. Verify Hash
	hash := sha256.Sum256([]byte(fullKey))
	inputHash := hex.EncodeToString(hash[:])

	if apiKey.KeyHash != inputHash {
		return nil, fmt.Errorf("invalid api key")
	}

	// 3. Check Expiry
	if !apiKey.ExpiresAt.IsZero() && time.Now().After(apiKey.ExpiresAt) {
		return nil, fmt.Errorf("api key expired")
	}

	// 4. Origin Validation for Publishable Keys
	if apiKey.Type == "publishable" {
		if len(apiKey.AllowedOrigins) > 0 {
			valid := false
			for _, o := range apiKey.AllowedOrigins {
				if o == "*" || o == currentOrigin {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("unauthorized origin: %s", currentOrigin)
			}
		}
	}

	// 5. Update Last Used (Async or non-blocking)
	go func() {
		_ = s.repo.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)
	}()

	return apiKey, nil
}

func (s *APIKeyService) ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error) {
	return s.repo.ListAPIKeys(ctx, tenantID)
}

func (s *APIKeyService) DeleteAPIKey(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	return s.repo.DeleteAPIKey(ctx, id, tenantID)
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random string: %w", err)
	}
	return hex.EncodeToString(b)[:length], nil
}
