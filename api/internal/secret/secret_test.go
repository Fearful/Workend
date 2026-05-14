package secret

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	raw, err := base64.StdEncoding.DecodeString(key)
	require.NoError(t, err)
	assert.Len(t, raw, keyLen)
}

func TestSealOpen(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	box, err := NewBox(key)
	require.NoError(t, err)

	plaintext := []byte("hello, workend!")
	blob, err := box.Seal(plaintext)
	require.NoError(t, err)
	assert.Greater(t, len(blob), len(plaintext))

	got, err := box.Open(blob)
	require.NoError(t, err)
	assert.Equal(t, plaintext, got)
}

func TestSealDifferentNonces(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	box, err := NewBox(key)
	require.NoError(t, err)

	plaintext := []byte("same input")
	blob1, err := box.Seal(plaintext)
	require.NoError(t, err)
	blob2, err := box.Seal(plaintext)
	require.NoError(t, err)

	assert.NotEqual(t, blob1, blob2)
}

func TestOpenWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	box1, _ := NewBox(key1)
	box2, _ := NewBox(key2)

	blob, err := box1.Seal([]byte("secret"))
	require.NoError(t, err)

	_, err = box2.Open(blob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decryption failed")
}

func TestOpenCorruptedBlob(t *testing.T) {
	key, _ := GenerateKey()
	box, _ := NewBox(key)

	blob, err := box.Seal([]byte("secret"))
	require.NoError(t, err)

	blob[len(blob)-1] ^= 0xff

	_, err = box.Open(blob)
	assert.Error(t, err)
}

func TestOpenTooShort(t *testing.T) {
	key, _ := GenerateKey()
	box, _ := NewBox(key)

	_, err := box.Open([]byte("short"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestNewBoxInvalidKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"not base64", "not-valid-base64!!!"},
		{"too short", base64.StdEncoding.EncodeToString([]byte("short"))},
		{"too long", base64.StdEncoding.EncodeToString(make([]byte, 64))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBox(tt.key)
			assert.Error(t, err)
		})
	}
}
