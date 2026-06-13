package domain

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
	maxFailedPIN = 5
)

type Credential struct {
	UserID      uuid.UUID
	PINHash     []byte
	FailedCount int8
	LockedUntil *time.Time
	UpdatedAt   time.Time
}

// HashPIN generates a fresh salt + argon2id hash, stored as "salt_hex:hash_hex".
func HashPIN(pin string) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	hash := argon2.IDKey([]byte(pin), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	encoded := fmt.Sprintf("%s:%s", hex.EncodeToString(salt), hex.EncodeToString(hash))
	return []byte(encoded), nil
}

// VerifyPIN checks a plaintext PIN against the stored "salt_hex:hash_hex" value.
func (c *Credential) VerifyPIN(pin string) bool {
	parts := strings.SplitN(string(c.PINHash), ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil || len(salt) != saltLen {
		return false
	}
	storedHash, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	computed := argon2.IDKey([]byte(pin), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return subtle.ConstantTimeCompare(computed, storedHash) == 1
}

func (c *Credential) IsLocked() bool {
	return c.LockedUntil != nil && time.Now().Before(*c.LockedUntil)
}

func (c *Credential) RecordFailedAttempt() {
	c.FailedCount++
	c.UpdatedAt = time.Now().UTC()
}

func (c *Credential) ShouldLock() bool {
	return c.FailedCount >= maxFailedPIN
}
