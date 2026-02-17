package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	keyLength     = 32
	saltLength    = 16
	nonceLength   = 12
	argon2Time    = 3
	argon2Memory  = 64 * 1024
	argon2Threads = 4
)

var (
	ErrInvalidKeyLength   = errors.New("invalid key length")
	ErrInvalidCiphertext  = errors.New("invalid ciphertext")
	ErrDecryptionFailed   = errors.New("decryption failed")
	ErrInvalidNonceLength = errors.New("invalid nonce length")
)

type Crypto struct {
	key []byte
}

func NewCrypto(masterPassword string, salt []byte) *Crypto {
	key := deriveKey(masterPassword, salt)
	return &Crypto{key: key}
}

func NewCryptoWithKey(key []byte) (*Crypto, error) {
	if len(key) != keyLength {
		return nil, ErrInvalidKeyLength
	}
	return &Crypto{key: key}, nil
}

func GenerateSalt() []byte {
	salt := make([]byte, saltLength)
	io.ReadFull(rand.Reader, salt)
	return salt
}

func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		argon2Time,
		argon2Memory,
		argon2Threads,
		keyLength,
	)
}

func (c *Crypto) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *Crypto) Decrypt(ciphertextB64 string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidNonceLength
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

func (c *Crypto) EncryptString(plaintext string) (string, error) {
	return c.Encrypt([]byte(plaintext))
}

func (c *Crypto) DecryptString(ciphertextB64 string) (string, error) {
	plaintext, err := c.Decrypt(ciphertextB64)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func ComputeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func VerifyChecksum(data []byte, expectedChecksum string) bool {
	return ComputeChecksum(data) == expectedChecksum
}

func (c *Crypto) Clear() {
	for i := range c.key {
		c.key[i] = 0
	}
}
