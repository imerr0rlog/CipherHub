package types

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
)

// Entry 表示密码库中的单个密码条目
type Entry struct {
	ID        string    `json:"id"`         // 唯一标识符
	Name      string    `json:"name"`       // 条目名称
	Username  string    `json:"username"`   // 用户名
	Password  string    `json:"password"`   // AES-256-GCM 加密，base64 编码
	URL       string    `json:"url,omitempty"` // 网站地址（可选）
	Notes     string    `json:"notes,omitempty"` // 备注（加密，base64 编码，可选）
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
	Tags      []string  `json:"tags,omitempty"` // 标签（可选）
}

// Vault 表示整个密码库结构
type Vault struct {
	Version   string            `json:"version"`   // 版本号
	Salt      string            `json:"salt"`      // Argon2 盐值，base64 编码
	Checksum  string            `json:"checksum"`  // SHA-256 完整性校验和
	Entries   []*Entry          `json:"entries"`   // 密码条目列表
	CreatedAt time.Time         `json:"created_at"` // 创建时间
	UpdatedAt time.Time         `json:"updated_at"` // 更新时间
	Metadata  map[string]string `json:"metadata,omitempty"` // 附加元数据（可选）
}

// StorageType 定义存储后端类型
type StorageType string

const (
	StorageTypeLocal  StorageType = "local"  // 本地文件存储
	StorageTypeWebDAV StorageType = "webdav" // WebDAV 云存储
)

// Config 保存应用程序的配置信息
type Config struct {
	DefaultStorage   StorageType   `json:"default_storage" yaml:"default_storage"` // 默认存储类型
	VaultPath        string        `json:"vault_path" yaml:"vault_path"`         // 密码库文件路径
	WebDAV           *WebDAVConfig `json:"webdav,omitempty" yaml:"webdav,omitempty"` // WebDAV 配置（可选）
	AutoSync         bool          `json:"auto_sync" yaml:"auto_sync"`           // 是否自动同步
	ClipboardTimeout int           `json:"clipboard_timeout" yaml:"clipboard_timeout"` // 剪贴板超时时间（秒）
}

// WebDAVConfig 定义 WebDAV 连接配置
type WebDAVConfig struct {
	URL                string `json:"url" yaml:"url"`                                   // WebDAV 服务器地址
	Username           string `json:"username" yaml:"username"`                         // 用户名
	Password           string `json:"password" yaml:"password"`                         // 密码
	RemotePath         string `json:"remote_path" yaml:"remote_path"`                   // 密码库在服务器上的路径
	ConfigRemotePath   string `json:"config_remote_path,omitempty" yaml:"config_remote_path,omitempty"` // 配置文件在服务器上的路径（可选）
	InsecureSkipVerify bool   `json:"insecure_skip_verify" yaml:"insecure_skip_verify"` // 是否跳过 TLS 证书验证
}

// NewVault 创建一个新的空密码库
func NewVault() *Vault {
	return &Vault{
		Version:   "1.0",
		Entries:   make([]*Entry, 0),
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewEntry 创建一个新的密码条目
func NewEntry(name string) (*Entry, error) {
	id, err := GenerateUUID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Entry{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      make([]string, 0),
	}, nil
}

// GenerateUUID 生成符合 RFC4122 标准的 UUID v4
func GenerateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b[:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:]), nil
}

// SecureRandomString 生成指定长度的加密安全随机字符串
func SecureRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

// getExeDir 获取程序可执行文件所在的目录
func getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

// DefaultConfig 返回应用程序的默认配置
func DefaultConfig() *Config {
	exeDir := getExeDir()
	return &Config{
		DefaultStorage:   StorageTypeLocal,
		VaultPath:        filepath.Join(exeDir, "vault.json"),
		AutoSync:         false,
		ClipboardTimeout: 30,
	}
}
