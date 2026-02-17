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
	ErrVaultNotOpen      = errors.New("vault: not open")
	ErrVaultAlreadyOpen  = errors.New("vault: already open")
	ErrEntryNotFound     = errors.New("vault: entry not found")
	ErrEntryExists       = errors.New("vault: entry already exists")
	ErrInvalidPassword   = errors.New("vault: invalid master password")
	ErrVaultCorrupted    = errors.New("vault: corrupted data")
)

type Manager struct {
	storage storage.Storage
	crypto  *crypto.Crypto
	vault   *types.Vault
	salt    []byte
	open    bool
}

func NewManager(storage storage.Storage) *Manager {
	return &Manager{
		storage: storage,
		open:    false,
	}
}

func (m *Manager) Init(masterPassword string) error {
	if m.storage.Exists() {
		return ErrVaultAlreadyOpen
	}

	m.salt = crypto.GenerateSalt()
	m.crypto = crypto.NewCrypto(masterPassword, m.salt)
	m.vault = types.NewVault()
	m.vault.Salt = base64.StdEncoding.EncodeToString(m.salt)
	m.open = true

	return m.save()
}

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

	data, err := json.MarshalIndent(m.vault, "", "  ")
	if err != nil {
		return err
	}

	m.vault.Checksum = crypto.ComputeChecksum(data)

	data, err = json.MarshalIndent(m.vault, "", "  ")
	if err != nil {
		return err
	}

	return m.storage.Write(data)
}

func (m *Manager) AddEntry(name, username, password, url, notes string, tags []string) (*types.Entry, error) {
	if !m.open {
		return nil, ErrVaultNotOpen
	}

	if m.findEntryByName(name) != nil {
		return nil, ErrEntryExists
	}

	entry := types.NewEntry(name)
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

func (m *Manager) ListEntries() ([]*types.Entry, error) {
	if !m.open {
		return nil, ErrVaultNotOpen
	}
	return m.vault.Entries, nil
}

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

func (m *Manager) SearchEntries(query string) ([]*types.Entry, error) {
	if !m.open {
		return nil, ErrVaultNotOpen
	}

	query = strings.ToLower(query)
	var results []*types.Entry

	for _, entry := range m.vault.Entries {
		if strings.Contains(strings.ToLower(entry.Name), query) ||
			strings.Contains(strings.ToLower(entry.Username), query) ||
			strings.Contains(strings.ToLower(entry.URL), query) {
			results = append(results, entry)
		}
		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, entry)
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

func (m *Manager) IsOpen() bool {
	return m.open
}

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

func (m *Manager) GeneratePassword(length int) (string, error) {
	return types.SecureRandomString(length), nil
}
