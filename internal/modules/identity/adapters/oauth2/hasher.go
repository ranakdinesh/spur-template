package oauth2

import (
	"context"
	"errors"

	"github.com/ory/fosite"
	"github.com/spurbase/spur/internal/platform/security"
)

type Argon2Hasher struct{}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{}
}

// Compare compares the hash with the data and returns an error if they don't match.
func (h *Argon2Hasher) Compare(ctx context.Context, hash, data []byte) error {
	match, err := security.VerifyPassword(string(data), string(hash))
	if err != nil {
		return err
	}
	if !match {
		return errors.New("password mismatch")
	}
	return nil
}

// Hash creates a hash from the data.
func (h *Argon2Hasher) Hash(ctx context.Context, data []byte) ([]byte, error) {
	// saltLength is ignored because security.HashPassword handles salt generation internally
	hash, err := security.HashPassword(string(data))
	if err != nil {
		return nil, err
	}
	return []byte(hash), nil
}

// Ensure interface compliance
var _ fosite.Hasher = (*Argon2Hasher)(nil)
