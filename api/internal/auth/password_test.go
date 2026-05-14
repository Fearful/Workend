package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAndVerify(t *testing.T) {
	password := "correct-horse-battery"
	hash, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	ok, err := VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerifyWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	require.NoError(t, err)

	ok, err := VerifyPassword("wrong-password", hash)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHashDifferentSalts(t *testing.T) {
	password := "same-password"
	hash1, err := HashPassword(password)
	require.NoError(t, err)
	hash2, err := HashPassword(password)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)

	ok1, _ := VerifyPassword(password, hash1)
	ok2, _ := VerifyPassword(password, hash2)
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func TestVerifyMalformedHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"garbage", "not-a-valid-hash"},
		{"wrong algorithm", "$bcrypt$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA"},
		{"too few parts", "$argon2id$v=19"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VerifyPassword("anything", tt.hash)
			assert.Error(t, err)
		})
	}
}
