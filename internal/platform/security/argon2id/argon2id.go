package argon2id

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

func DefaultParams() Params {
	return Params{
		Time:    2,
		Memory:  64 * 1024, // 64 MB
		Threads: 1,
		KeyLen:  32,
		SaltLen: 16,
	}
}

type Hasher struct {
	p Params
}

func New(p Params) *Hasher { return &Hasher{p: p} }

func (h *Hasher) Hash(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty password")
	}

	salt := make([]byte, h.p.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(raw), salt, h.p.Time, h.p.Memory, h.p.Threads, h.p.KeyLen)

	// Format:
	// $argon2id$v=19$m=65536,t=2,p=1$<salt_b64>$<key_b64>
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	keyB64 := base64.RawStdEncoding.EncodeToString(key)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.p.Memory, h.p.Time, h.p.Threads, saltB64, keyB64,
	), nil
}

func (h *Hasher) Compare(encodedHash, raw string) bool {
	p, salt, key, err := parse(encodedHash)
	if err != nil {
		return false
	}
	rawKey := argon2.IDKey([]byte(raw), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	return subtle.ConstantTimeCompare(rawKey, key) == 1
}

func parse(encoded string) (Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Params{}, nil, nil, errors.New("invalid hash format")
	}

	var mem uint32
	var time uint32
	var threads uint8

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &time, &threads)
	if err != nil {
		return Params{}, nil, nil, errors.New("invalid params")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, errors.New("invalid salt")
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Params{}, nil, nil, errors.New("invalid key")
	}

	return Params{
		Time:    time,
		Memory:  mem,
		Threads: threads,
		KeyLen:  uint32(len(key)),
		SaltLen: uint32(len(salt)),
	}, salt, key, nil
}
