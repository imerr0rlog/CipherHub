package types

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // AES-256-GCM encrypted, base64 encoded
	URL       string    `json:"url,omitempty"`
	Notes     string    `json:"notes,omitempty"` // Encrypted, base64 encoded
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      []string  `json:"tags,omitempty"`
}

type Vault struct {
	Version   string            `json:"version"`
	Salt      string            `json:"salt"`     // Argon2 salt, base64 encoded
	Checksum  string            `json:"checksum"` // SHA-256 for integrity verification
	Entries   []*Entry          `json:"entries"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type StorageType string

const (
	StorageTypeLocal  StorageType = "local"
	StorageTypeWebDAV StorageType = "webdav"
)

type Config struct {
	DefaultStorage   StorageType   `json:"default_storage" yaml:"default_storage"`
	VaultPath        string        `json:"vault_path" yaml:"vault_path"`
	WebDAV           *WebDAVConfig `json:"webdav,omitempty" yaml:"webdav,omitempty"`
	AutoSync         bool          `json:"auto_sync" yaml:"auto_sync"`
	ClipboardTimeout int           `json:"clipboard_timeout" yaml:"clipboard_timeout"`
}

type WebDAVConfig struct {
	URL                string `json:"url" yaml:"url"`
	Username           string `json:"username" yaml:"username"`
	Password           string `json:"password" yaml:"password"` // Encrypted
	RemotePath         string `json:"remote_path" yaml:"remote_path"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
}

func NewVault() *Vault {
	return &Vault{
		Version:   "1.0",
		Entries:   make([]*Entry, 0),
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewEntry(name string) *Entry {
	now := time.Now()
	return &Entry{
		ID:        GenerateUUID(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      make([]string, 0),
	}
}

func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b[:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:])
}

func SecureRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

func DefaultConfig() *Config {
	exeDir := getExeDir()
	return &Config{
		DefaultStorage:   StorageTypeLocal,
		VaultPath:        filepath.Join(exeDir, "vault.json"),
		AutoSync:         false,
		ClipboardTimeout: 30,
	}
}
