package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// ConfigCrypto handles AES-256-GCM encryption for config values.
type ConfigCrypto struct {
	key []byte
}

// NewConfigCrypto creates a new ConfigCrypto with a key derived from the passphrase.
func NewConfigCrypto(passphrase string) *ConfigCrypto {
	hash := sha256.Sum256([]byte(passphrase))
	return &ConfigCrypto{key: hash[:]}
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext.
func (c *ConfigCrypto) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded AES-256-GCM ciphertext.
func (c *ConfigCrypto) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Mask returns a masked version of a value (first 3 + ••• + last 3).
func Mask(value string) string {
	if len(value) <= 6 {
		return "**"
	}
	return value[:3] + "•••" + value[len(value)-3:]
}
