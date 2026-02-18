package vault

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/imerr0rlog/CipherHub/internal/crypto"
	"github.com/imerr0rlog/CipherHub/internal/storage"
	"github.com/imerr0rlog/CipherHub/pkg/types"
)

var (
	// ErrVaultNotOpen 表示密码库未打开时尝试执行操作
	ErrVaultNotOpen      = errors.New("vault: not open")
	// ErrVaultAlreadyOpen 表示尝试打开已打开的密码库
	ErrVaultAlreadyOpen  = errors.New("vault: already open")
	// ErrVaultExists 表示尝试初始化已存在的密码库
	ErrVaultExists       = errors.New("vault: already exists")
	// ErrEntryNotFound 表示未找到指定的密码条目
	ErrEntryNotFound     = errors.New("vault: entry not found")
	// ErrEntryExists 表示尝试添加已存在的密码条目
	ErrEntryExists       = errors.New("vault: entry already exists")
	// ErrInvalidPassword 表示主密码不正确
	ErrInvalidPassword   = errors.New("vault: invalid master password")
	// ErrVaultCorrupted 表示密码库数据已损坏
	ErrVaultCorrupted    = errors.New("vault: corrupted data")
	// ErrRandomGenFailed 表示随机数生成失败
	ErrRandomGenFailed   = errors.New("vault: random generation failed")
)

// Manager 负责密码库的所有操作，包括初始化、打开、关闭密码库，以及密码条目的增删改查
type Manager struct {
	storage storage.Storage
	crypto  *crypto.Crypto
	vault   *types.Vault
	salt    []byte
	open    bool
}

// NewManager 创建一个新的密码库管理器实例
//
// 参数:
//   storage - 存储接口，用于读写密码库数据
//
// 返回:
//   新的 Manager 实例
func NewManager(storage storage.Storage) *Manager {
	return &Manager{
		storage: storage,
		open:    false,
	}
}

// Init 初始化一个新的密码库
//
// 参数:
//   masterPassword - 主密码，用于加密密码库
//
// 返回:
//   成功时返回 nil，失败时返回相应的错误
func (m *Manager) Init(masterPassword string) error {
	if m.storage.Exists() {
		return ErrVaultExists
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return ErrRandomGenFailed
	}
	m.salt = salt
	m.crypto = crypto.NewCrypto(masterPassword, m.salt)
	m.vault = types.NewVault()
	m.vault.Salt = base64.StdEncoding.EncodeToString(m.salt)
	m.open = true

	return m.save()
}

// Open 打开现有的密码库
//
// 参数:
//   masterPassword - 主密码，用于解密密码库
//
// 返回:
//   成功时返回 nil，失败时返回相应的错误
func (m *Manager) Open(masterPassword string) error {
	if m.open {
		return ErrVaultAlreadyOpen
	}

	data, err := m.storage.Read()
	if err != nil {
		return err
	}

	var vault types.Vault
	if err := json.Unmarshal(data, &vault); err != nil {
		return ErrVaultCorrupted
	}

	salt, err := base64.StdEncoding.DecodeString(vault.Salt)
	if err != nil {
		return ErrVaultCorrupted
	}

	m.salt = salt
	m.crypto = crypto.NewCrypto(masterPassword, salt)
	m.vault = &vault
	m.open = true

	return nil
}

// Close 关闭密码库，清空内存中的敏感数据
func (m *Manager) Close() {
	if m.crypto != nil {
		m.crypto.Clear()
	}
	m.crypto = nil
	m.vault = nil
	m.salt = nil
	m.open = false
}

func (m *Manager) save() error {
	if !m.open {
		return ErrVaultNotOpen
	}

	m.vault.UpdatedAt = time.Now()

	tempChecksum := m.vault.Checksum
	m.vault.Checksum = ""

	data, err := json.MarshalIndent(m.vault, "", "  ")
	if err != nil {
		m.vault.Checksum = tempChecksum
		return err
	}

	m.vault.Checksum = crypto.ComputeChecksum(data)

	data, err = json.MarshalIndent(m.vault, "", "  ")
	if err != nil {
		return err
	}

	return m.storage.Write(data)
}

// AddEntry 添加新的密码条目
//
// 参数:
//   name - 条目的唯一名称
//   username - 用户名
//   password - 密码（会被加密存储）
//   url - 网站地址
//   notes - 备注信息（会被加密存储）
//   tags - 标签列表
//
// 返回:
//   新创建的密码条目和可能的错误
func (m *Manager) AddEntry(name, username, password, url, notes string, tags []string) (*types.Entry, error) {
	if !m.open {
		return nil, ErrVaultNotOpen
	}

	if m.findEntryByName(name) != nil {
		return nil, ErrEntryExists
	}

	entry, err := types.NewEntry(name)
	if err != nil {
		return nil, ErrRandomGenFailed
	}
	entry.Username = username

	encPassword, err := m.crypto.EncryptString(password)
	if err != nil {
		return nil, err
	}
	entry.Password = encPassword

	entry.URL = url

	if notes != "" {
		encNotes, err := m.crypto.EncryptString(notes)
		if err != nil {
			return nil, err
		}
		entry.Notes = encNotes
	}

	entry.Tags = tags

	m.vault.Entries = append(m.vault.Entries, entry)
	if err := m.save(); err != nil {
		return nil, err
	}

	return entry, nil
}

// GetEntry 根据名称获取密码条目（密码和备注仍保持加密状态）
//
// 参数:
//   name - 条目名称
//
// 返回:
//   找到的密码条目和可能的错误
func (m *Manager) GetEntry(name string) (*types.Entry, error) {
	if !m.open {
		return nil, ErrVaultNotOpen
	}

	entry := m.findEntryByName(name)
	if entry == nil {
		return nil, ErrEntryNotFound
	}

	return entry, nil
}

// GetDecryptedPassword 获取解密后的密码
//
// 参数:
//   name - 条目名称
//
// 返回:
//   解密后的密码和可能的错误
func (m *Manager) GetDecryptedPassword(name string) (string, error) {
	if !m.open {
		return "", ErrVaultNotOpen
	}

	entry := m.findEntryByName(name)
	if entry == nil {
		return "", ErrEntryNotFound
	}

	return m.crypto.DecryptString(entry.Password)
}

// GetDecryptedNotes 获取解密后的备注
//
// 参数:
//   name - 条目名称
//
// 返回:
//   解密后的备注和可能的错误
func (m *Manager) GetDecryptedNotes(name string) (string, error) {
	if !m.open {
		return "", ErrVaultNotOpen
	}

	entry := m.findEntryByName(name)
	if entry == nil {
		return "", ErrEntryNotFound
	}

	if entry.Notes == "" {
		return "", nil
	}

	return m.crypto.DecryptString(entry.Notes)
}

// ListEntries 列出所有密码条目
//
// 返回:
//   所有密码条目列表和可能的错误
func (m *Manager) ListEntries() ([]*types.Entry, error) {
	if !m.open {
		return nil, ErrVaultNotOpen
	}
	return m.vault.Entries, nil
}

// UpdateEntry 更新密码条目
//
// 参数:
//   name - 要更新的条目名称
//   updates - 包含要更新字段的映射，支持的键: username, password, url, notes
//
// 返回:
//   更新后的密码条目和可能的错误
func (m *Manager) UpdateEntry(name string, updates map[string]string) (*types.Entry, error) {
	if !m.open {
		return nil, ErrVaultNotOpen
	}

	entry := m.findEntryByName(name)
	if entry == nil {
		return nil, ErrEntryNotFound
	}

	if username, ok := updates["username"]; ok {
		entry.Username = username
	}
	if password, ok := updates["password"]; ok {
		encPassword, err := m.crypto.EncryptString(password)
		if err != nil {
			return nil, err
		}
		entry.Password = encPassword
	}
	if url, ok := updates["url"]; ok {
		entry.URL = url
	}
	if notes, ok := updates["notes"]; ok {
		if notes == "" {
			entry.Notes = ""
		} else {
			encNotes, err := m.crypto.EncryptString(notes)
			if err != nil {
				return nil, err
			}
			entry.Notes = encNotes
		}
	}

	entry.UpdatedAt = time.Now()
	if err := m.save(); err != nil {
		return nil, err
	}

	return entry, nil
}

// DeleteEntry 删除密码条目
//
// 参数:
//   name - 要删除的条目名称
//
// 返回:
//   可能的错误
func (m *Manager) DeleteEntry(name string) error {
	if !m.open {
		return ErrVaultNotOpen
	}

	idx := m.findEntryIndexByName(name)
	if idx == -1 {
		return ErrEntryNotFound
	}

	m.vault.Entries = append(m.vault.Entries[:idx], m.vault.Entries[idx+1:]...)
	return m.save()
}

// SearchEntries 搜索密码条目，按名称、用户名、URL或标签进行搜索
//
// 参数:
//   query - 搜索关键词（不区分大小写）
//
// 返回:
//   匹配的密码条目列表和可能的错误
func (m *Manager) SearchEntries(query string) ([]*types.Entry, error) {
	if !m.open {
		return nil, ErrVaultNotOpen
	}

	query = strings.ToLower(query)
	seen := make(map[string]bool)
	var results []*types.Entry

	for _, entry := range m.vault.Entries {
		if seen[entry.ID] {
			continue
		}
		if strings.Contains(strings.ToLower(entry.Name), query) ||
			strings.Contains(strings.ToLower(entry.Username), query) ||
			strings.Contains(strings.ToLower(entry.URL), query) {
			results = append(results, entry)
			seen[entry.ID] = true
			continue
		}
		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, entry)
				seen[entry.ID] = true
				break
			}
		}
	}

	return results, nil
}

func (m *Manager) findEntryByName(name string) *types.Entry {
	for _, entry := range m.vault.Entries {
		if entry.Name == name {
			return entry
		}
	}
	return nil
}

func (m *Manager) findEntryIndexByName(name string) int {
	for i, entry := range m.vault.Entries {
		if entry.Name == name {
			return i
		}
	}
	return -1
}

// IsOpen 检查密码库是否已打开
//
// 返回:
//   密码库已打开时返回 true，否则返回 false
func (m *Manager) IsOpen() bool {
	return m.open
}

// VaultInfo 获取密码库的基本信息
//
// 返回:
//   包含密码库信息的映射，包括 open（是否打开）、version（版本）、entries（条目数量）、created_at（创建时间）、updated_at（更新时间）
func (m *Manager) VaultInfo() map[string]interface{} {
	if !m.open {
		return map[string]interface{}{"open": false}
	}

	return map[string]interface{}{
		"open":       true,
		"version":    m.vault.Version,
		"entries":    len(m.vault.Entries),
		"created_at": m.vault.CreatedAt,
		"updated_at": m.vault.UpdatedAt,
	}
}

// Sync 将当前密码库同步到远程存储
//
// 参数:
//   remote - 远程存储接口
//
// 返回:
//   可能的错误
func (m *Manager) Sync(remote storage.Storage) error {
	if !m.open {
		return ErrVaultNotOpen
	}

	data, err := json.MarshalIndent(m.vault, "", "  ")
	if err != nil {
		return err
	}

	return remote.Write(data)
}

// Pull 从远程存储拉取密码库并替换本地密码库
//
// 参数:
//   remote - 远程存储接口
//   masterPassword - 主密码，用于解密拉取的密码库
//
// 返回:
//   可能的错误
func (m *Manager) Pull(remote storage.Storage, masterPassword string) error {
	data, err := remote.Read()
	if err != nil {
		return err
	}

	var vault types.Vault
	if err := json.Unmarshal(data, &vault); err != nil {
		return ErrVaultCorrupted
	}

	salt, err := base64.StdEncoding.DecodeString(vault.Salt)
	if err != nil {
		return ErrVaultCorrupted
	}

	if m.open {
		m.Close()
	}

	m.salt = salt
	m.crypto = crypto.NewCrypto(masterPassword, salt)
	m.vault = &vault
	m.open = true

	return nil
}

// GeneratePassword 生成安全的随机密码
//
// 参数:
//   length - 密码长度
//
// 返回:
//   生成的密码和可能的错误
func (m *Manager) GeneratePassword(length int) (string, error) {
	return types.SecureRandomString(length)
}
