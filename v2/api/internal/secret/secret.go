// Package secret provides symmetric encryption for sensitive values stored
// at rest (e.g., third-party OAuth tokens). Uses NaCl secretbox; the key is
// 32 bytes loaded from the WORKEND_TOKEN_KEY env var (base64-encoded).
package secret

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/nacl/secretbox"
)

const nonceLen = 24
const keyLen = 32

type Box struct {
	key [keyLen]byte
}

// NewBox parses a base64-encoded 32-byte key.
func NewBox(b64Key string) (*Box, error) {
	raw, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}
	if len(raw) != keyLen {
		return nil, fmt.Errorf("token key must be %d bytes (got %d)", keyLen, len(raw))
	}
	b := &Box{}
	copy(b.key[:], raw)
	return b, nil
}

// Seal encrypts plaintext, returning nonce||ciphertext as a single blob
// suitable for storing in a BYTEA column.
func (b *Box) Seal(plaintext []byte) ([]byte, error) {
	var nonce [nonceLen]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	out := make([]byte, nonceLen)
	copy(out, nonce[:])
	out = secretbox.Seal(out, plaintext, &nonce, &b.key)
	return out, nil
}

// Open decrypts a blob produced by Seal.
func (b *Box) Open(blob []byte) ([]byte, error) {
	if len(blob) < nonceLen+secretbox.Overhead {
		return nil, errors.New("ciphertext too short")
	}
	var nonce [nonceLen]byte
	copy(nonce[:], blob[:nonceLen])
	out, ok := secretbox.Open(nil, blob[nonceLen:], &nonce, &b.key)
	if !ok {
		return nil, errors.New("decryption failed")
	}
	return out, nil
}

// GenerateKey returns a fresh base64-encoded 32-byte key suitable for the
// WORKEND_TOKEN_KEY env var. Useful for `workend keygen` scaffolding.
func GenerateKey() (string, error) {
	b := make([]byte, keyLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
