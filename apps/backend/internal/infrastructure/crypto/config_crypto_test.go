package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	key := "test-encryption-key-32-bytes!!"
	c := NewConfigCrypto(key)

	plaintext := "sk-abc123def456ghi789"
	encrypted, err := c.Encrypt(plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, encrypted)

	decrypted, err := c.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecrypt_DifferentCiphertexts(t *testing.T) {
	key := "test-encryption-key-32-bytes!!"
	c := NewConfigCrypto(key)

	plaintext := "my-api-key"
	enc1, _ := c.Encrypt(plaintext)
	enc2, _ := c.Encrypt(plaintext)

	// Same plaintext should produce different ciphertexts (random nonce)
	assert.NotEqual(t, enc1, enc2)

	// But both should decrypt to the same plaintext
	dec1, _ := c.Decrypt(enc1)
	dec2, _ := c.Decrypt(enc2)
	assert.Equal(t, dec1, dec2)
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	key := "test-encryption-key-32-bytes!!"
	c := NewConfigCrypto(key)

	_, err := c.Decrypt("not-valid-base64-ciphertext!!!")
	assert.Error(t, err)
}

func TestMask(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-abc123def456ghi789jkl", "sk-•••jkl"},
		{"short", "**"},
		{"ab", "**"},
		{"", "**"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, Mask(tt.input))
		})
	}
}
