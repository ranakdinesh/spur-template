package oauth2

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/rs/zerolog"
)

// NewProvider creates a configured Fosite OAuth2 provider.
// secretKey, rsaKeyPath, and issuer all come from identity.Config — never hardcoded.
func NewProvider(storage fosite.Storage, logger zerolog.Logger, secretKey, rsaKeyPath, issuer string) (fosite.OAuth2Provider, error) {
	config := &fosite.Config{
		AccessTokenLifespan:  24 * time.Hour,
		RefreshTokenLifespan: 30 * 24 * time.Hour,
		GlobalSecret:         []byte(secretKey),
		RefreshTokenScopes:   []string{"offline", "offline_access"},
		ClientSecretsHasher:  NewArgon2Hasher(),
		IDTokenIssuer:        issuer, // from config, not hardcoded
	}

	privateKey, err := LoadPrivateKey(rsaKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load RSA key: %w", err)
	}

	keyGetter := func(ctx context.Context) (interface{}, error) {
		return privateKey, nil
	}

	hmacStrategy := compose.NewOAuth2HMACStrategy(config)
	jwtStrategy  := compose.NewOAuth2JWTStrategy(keyGetter, hmacStrategy, config)
	oidcStrategy := compose.NewOpenIDConnectStrategy(keyGetter, config)

	return compose.Compose(
		config,
		storage,
		&compose.CommonStrategy{
			CoreStrategy:               jwtStrategy,
			OpenIDConnectTokenStrategy: oidcStrategy,
		},
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OpenIDConnectExplicitFactory,
		compose.OpenIDConnectRefreshFactory,
		compose.OAuth2PKCEFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
	), nil
}

// LoadPrivateKey reads an RSA private key from path.
// If the file does not exist, a new 2048-bit key is generated and written
// with 0600 permissions (owner read/write only).
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return nil, fmt.Errorf("RSA key path is empty")
	}

	b, err := os.ReadFile(path)
	if err == nil {
		block, _ := pem.Decode(b)
		if block != nil {
			if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				return key, nil
			}
			if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
				if rsaKey, ok := key.(*rsa.PrivateKey); ok {
					return rsaKey, nil
				}
			}
		}
	}

	// Generate new key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}

	// FIX: 0600 — owner read/write only, never world-readable
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("create key file %s: %w", path, err)
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return nil, fmt.Errorf("encode RSA key: %w", err)
	}

	return key, nil
}
