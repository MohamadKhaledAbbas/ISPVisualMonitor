package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		cost        int
		expectError bool
		errorType   error
	}{
		{
			name:        "valid password with default cost",
			password:    "mySecurePassword123!",
			cost:        DefaultBcryptCost,
			expectError: false,
		},
		{
			name:        "valid password with custom cost",
			password:    "anotherPassword456",
			cost:        10,
			expectError: false,
		},
		{
			name:        "password too short",
			password:    "short",
			cost:        DefaultBcryptCost,
			expectError: true,
			errorType:   ErrPasswordTooShort,
		},
		{
			name:        "minimum length password",
			password:    "12345678",
			cost:        DefaultBcryptCost,
			expectError: false,
		},
		{
			name:        "invalid cost should use default",
			password:    "validPassword123",
			cost:        100, // Invalid cost
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password, tt.cost)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if hash == "" {
				t.Error("expected non-empty hash")
			}

			// Verify the hash is valid bcrypt
			if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
				t.Error("hash does not appear to be valid bcrypt")
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testPassword123!"
	hash, err := HashPassword(password, DefaultBcryptCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name        string
		password    string
		hash        string
		expectError bool
		errorType   error
	}{
		{
			name:        "correct password",
			password:    password,
			hash:        hash,
			expectError: false,
		},
		{
			name:        "incorrect password",
			password:    "wrongPassword",
			hash:        hash,
			expectError: true,
			errorType:   ErrPasswordMismatch,
		},
		{
			name:        "empty password",
			password:    "",
			hash:        hash,
			expectError: true,
			errorType:   ErrPasswordMismatch,
		},
		{
			name:        "invalid hash",
			password:    password,
			hash:        "invalid-hash",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.password, tt.hash)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("expected error %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyPasswordSafe(t *testing.T) {
	password := "testPassword123!"
	hash, err := HashPassword(password, DefaultBcryptCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "incorrect password",
			password: "wrongPassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "invalid hash",
			password: password,
			hash:     "invalid-hash",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPasswordSafe(tt.password, tt.hash)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCompareHashes(t *testing.T) {
	password := "testPassword123!"
	hash1, _ := HashPassword(password, 10)
	hash2, _ := HashPassword(password, 10)

	tests := []struct {
		name     string
		hash1    string
		hash2    string
		expected bool
	}{
		{
			name:     "identical hashes",
			hash1:    hash1,
			hash2:    hash1,
			expected: true,
		},
		{
			name:     "different hashes same password",
			hash1:    hash1,
			hash2:    hash2,
			expected: false,
		},
		{
			name:     "completely different hashes",
			hash1:    hash1,
			hash2:    "different",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareHashes(tt.hash1, tt.hash2)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Benchmark password hashing
func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkPassword123!"
	costs := []int{10, 12, 14}

	for _, cost := range costs {
		b.Run("cost="+string(rune(cost+'0')), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = HashPassword(password, cost)
			}
		})
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkPassword123!"
	hash, _ := HashPassword(password, DefaultBcryptCost)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyPassword(password, hash)
	}
}

// Test that hashing produces different hashes for same password
func TestHashUniqueness(t *testing.T) {
	password := "testPassword123!"
	cost := DefaultBcryptCost

	hash1, err1 := HashPassword(password, cost)
	hash2, err2 := HashPassword(password, cost)

	if err1 != nil || err2 != nil {
		t.Fatalf("failed to hash passwords: %v, %v", err1, err2)
	}

	if hash1 == hash2 {
		t.Error("expected different hashes for same password due to salt")
	}

	// But both should verify correctly
	if err := VerifyPassword(password, hash1); err != nil {
		t.Errorf("hash1 verification failed: %v", err)
	}

	if err := VerifyPassword(password, hash2); err != nil {
		t.Errorf("hash2 verification failed: %v", err)
	}
}

// Test various password lengths
func TestPasswordLengths(t *testing.T) {
	tests := []struct {
		length      int
		expectError bool
	}{
		{0, true},
		{1, true},
		{7, true},
		{8, false},
		{16, false},
		{32, false},
		{64, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			password := strings.Repeat("a", tt.length)
			_, err := HashPassword(password, DefaultBcryptCost)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Test cost validation
func TestCostValidation(t *testing.T) {
	password := "validPassword123"

	tests := []struct {
		name string
		cost int
	}{
		{"below minimum", bcrypt.MinCost - 1},
		{"minimum", bcrypt.MinCost},
		{"default", DefaultBcryptCost},
		{"maximum", bcrypt.MaxCost},
		{"above maximum", bcrypt.MaxCost + 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(password, tt.cost)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Should still produce valid hash (using default if invalid)
			if hash == "" {
				t.Error("expected non-empty hash")
			}

			// Verify it works
			if err := VerifyPassword(password, hash); err != nil {
				t.Errorf("verification failed: %v", err)
			}
		})
	}
}
