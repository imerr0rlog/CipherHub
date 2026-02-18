// Package crypto 提供密码加密和解密功能，使用 AES-256-GCM 算法和 Argon2id 密钥派生
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
	keyLength     = 32          // AES-256 需要 32 字节密钥
	saltLength    = 16          // Argon2 盐值长度
	nonceLength   = 12          // GCM 推荐的 nonce 长度
	argon2Time    = 3           // Argon2 迭代次数
	argon2Memory  = 64 * 1024   // Argon2 内存使用量 (KB)
	argon2Threads = 4           // Argon2 并行线程数
)

var (
	// ErrInvalidKeyLength 表示密钥长度无效
	ErrInvalidKeyLength = errors.New("invalid key length")
	// ErrInvalidCiphertext 表示密文格式无效
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrDecryptionFailed 表示解密失败
	ErrDecryptionFailed = errors.New("decryption failed")
	// ErrInvalidNonceLength 表示 nonce 长度无效
	ErrInvalidNonceLength = errors.New("invalid nonce length")
)

// Crypto 负责加密和解密操作，使用 AES-256-GCM 算法
type Crypto struct {
	key []byte
}

// NewCrypto 使用主密码和盐值创建一个新的加密实例
func NewCrypto(masterPassword string, salt []byte) *Crypto {
	key := deriveKey(masterPassword, salt)
	return &Crypto{key: key}
}

// NewCryptoWithKey 使用直接提供的密钥创建加密实例
func NewCryptoWithKey(key []byte) (*Crypto, error) {
	if len(key) != keyLength {
		return nil, ErrInvalidKeyLength
	}
	return &Crypto{key: key}, nil
}

// GenerateSalt 生成加密安全的随机盐值
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// deriveKey 使用 Argon2id 算法从主密码派生加密密钥
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

// Encrypt 使用 AES-256-GCM 加密字节数据
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
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 base64 编码的密文
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

// EncryptString 加密字符串
func (c *Crypto) EncryptString(plaintext string) (string, error) {
	return c.Encrypt([]byte(plaintext))
}

// DecryptString 解密为字符串
func (c *Crypto) DecryptString(ciphertextB64 string) (string, error) {
	plaintext, err := c.Decrypt(ciphertextB64)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// ComputeChecksum 计算数据的 SHA-256 校验和
func ComputeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// VerifyChecksum 验证数据的校验和是否匹配
func VerifyChecksum(data []byte, expectedChecksum string) bool {
	return ComputeChecksum(data) == expectedChecksum
}

// Clear 安全清除内存中的密钥数据
func (c *Crypto) Clear() {
	for i := range c.key {
		c.key[i] = 0
	}
}
