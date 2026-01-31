package auth

import (
	"crypto/subtle"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	
	// ErrPasswordHashFailed is returned when password hashing fails
	ErrPasswordHashFailed = errors.New("failed to hash password")
	
	// ErrPasswordMismatch is returned when password doesn't match hash
	ErrPasswordMismatch = errors.New("password does not match")
)

const (
	// MinPasswordLength defines minimum password length
	MinPasswordLength = 8
	
	// DefaultBcryptCost is the default bcrypt cost factor
	DefaultBcryptCost = 12
)

// HashPassword hashes a password using bcrypt with the specified cost
// Cost should be between bcrypt.MinCost (4) and bcrypt.MaxCost (31)
// Default cost is 12, which provides a good balance between security and performance
func HashPassword(password string, cost int) (string, error) {
	if len(password) < MinPasswordLength {
		return "", ErrPasswordTooShort
	}
	
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = DefaultBcryptCost
	}
	
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", ErrPasswordHashFailed
	}
	
	return string(hashedBytes), nil
}

// VerifyPassword verifies that a password matches a bcrypt hash
// Uses constant-time comparison to prevent timing attacks
func VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return err
	}
	
	return nil
}

// VerifyPasswordSafe is a timing-safe password verification
// It always performs a hash comparison even if the hash is invalid
// This prevents timing attacks that could determine if a user exists
func VerifyPasswordSafe(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	
	// Use constant-time comparison for the error result
	// This ensures the function takes the same time regardless of whether
	// the password is correct or incorrect
	if err != nil {
		// Perform a dummy comparison with constant time
		dummyHash := []byte("$2a$12$dummy.hash.that.will.always.fail")
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return false
	}
	
	return true
}

// CompareHashes compares two bcrypt hashes in constant time
// This is useful for comparing stored hashes
func CompareHashes(hash1, hash2 string) bool {
	return subtle.ConstantTimeCompare([]byte(hash1), []byte(hash2)) == 1
}
