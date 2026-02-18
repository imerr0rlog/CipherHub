package api

import (
	"encoding/json"
	"os"

	"github.com/imerr0rlog/CipherHub/internal/crypto"
	"github.com/imerr0rlog/CipherHub/internal/storage"
	"github.com/imerr0rlog/CipherHub/internal/vault"
	"github.com/imerr0rlog/CipherHub/pkg/types"
)

type Client struct {
	manager    *vault.Manager
	storage    storage.Storage
	config     *types.Config
	configPath string
}

type ClientOptions struct {
	VaultPath  string
	ConfigPath string
}

func NewClient(cfg *types.Config) (*Client, error) {
	st, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		storage: st,
		config:  cfg,
		manager: vault.NewManager(st),
	}, nil
}

func NewClientWithOptions(opts *ClientOptions) (*Client, error) {
	cfg := types.DefaultConfig()
	configPath := opts.ConfigPath

	if configPath == "" {
		exePath, err := os.Executable()
		if err != nil {
			return nil, err
		}
		configPath = exePath + ".json"
	}

	if opts.VaultPath != "" {
		cfg.VaultPath = opts.VaultPath
	}

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	client.configPath = configPath
	return client, nil
}

func LoadConfig(path string) (*types.Config, error) {
	return storage.LoadConfig(path)
}

func SaveConfig(path string, cfg *types.Config) error {
	return storage.SaveConfig(path, cfg)
}

func DefaultConfig() *types.Config {
	return types.DefaultConfig()
}

func (c *Client) Config() *types.Config {
	return c.config
}

func (c *Client) ConfigPath() string {
	return c.configPath
}

func (c *Client) SetVaultPath(path string) {
	c.config.VaultPath = path
	c.storage = storage.NewLocalStorage(path)
	c.manager = vault.NewManager(c.storage)
}

func (c *Client) InitVault(masterPassword string) error {
	return c.manager.Init(masterPassword)
}

func (c *Client) OpenVault(masterPassword string) error {
	return c.manager.Open(masterPassword)
}

func (c *Client) CloseVault() {
	c.manager.Close()
}

func (c *Client) IsVaultOpen() bool {
	return c.manager.IsOpen()
}

func (c *Client) VaultExists() bool {
	return c.storage.Exists()
}

func (c *Client) AddEntry(name, username, password, url, notes string, tags []string) (*types.Entry, error) {
	return c.manager.AddEntry(name, username, password, url, notes, tags)
}

func (c *Client) GetEntry(name string) (*types.Entry, error) {
	return c.manager.GetEntry(name)
}

func (c *Client) GetDecryptedPassword(name string) (string, error) {
	return c.manager.GetDecryptedPassword(name)
}

func (c *Client) GetDecryptedNotes(name string) (string, error) {
	return c.manager.GetDecryptedNotes(name)
}

func (c *Client) ListEntries() ([]*types.Entry, error) {
	return c.manager.ListEntries()
}

func (c *Client) SearchEntries(query string) ([]*types.Entry, error) {
	return c.manager.SearchEntries(query)
}

func (c *Client) UpdateEntry(name string, updates map[string]string) (*types.Entry, error) {
	return c.manager.UpdateEntry(name, updates)
}

func (c *Client) DeleteEntry(name string) error {
	return c.manager.DeleteEntry(name)
}

func (c *Client) GeneratePassword(length int) string {
	return types.SecureRandomString(length)
}

func (c *Client) VaultInfo() map[string]interface{} {
	return c.manager.VaultInfo()
}

type SyncOptions struct {
	SyncVault  bool
	SyncConfig bool
}

func (c *Client) Sync(remote storage.Storage) error {
	return c.manager.Sync(remote)
}

func (c *Client) Pull(remote storage.Storage, masterPassword string) error {
	return c.manager.Pull(remote, masterPassword)
}

func (c *Client) SyncToWebDAV(opts *SyncOptions) error {
	if c.config.WebDAV == nil || c.config.WebDAV.URL == "" {
		return ErrWebDAVNotConfigured
	}

	webdavStorage := storage.NewWebDAVStorage(c.config.WebDAV)
	if err := webdavStorage.Connect(); err != nil {
		return err
	}

	syncVault := opts == nil || opts.SyncVault
	syncConfig := opts == nil || opts.SyncConfig

	if syncVault {
		if err := c.manager.Sync(webdavStorage); err != nil {
			return err
		}
	}

	if syncConfig && c.config.WebDAV.ConfigRemotePath != "" && c.configPath != "" {
		configData, err := os.ReadFile(c.configPath)
		if err != nil {
			return err
		}
		configStorage := storage.NewWebDAVStorage(&types.WebDAVConfig{
			URL:                c.config.WebDAV.URL,
			Username:           c.config.WebDAV.Username,
			Password:           c.config.WebDAV.Password,
			RemotePath:         c.config.WebDAV.ConfigRemotePath,
			InsecureSkipVerify: c.config.WebDAV.InsecureSkipVerify,
		})
		if err := configStorage.Connect(); err != nil {
			return err
		}
		if err := configStorage.Write(configData); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) PullFromWebDAV(opts *SyncOptions) error {
	if c.config.WebDAV == nil || c.config.WebDAV.URL == "" {
		return ErrWebDAVNotConfigured
	}

	syncVault := opts == nil || opts.SyncVault
	syncConfig := opts == nil || opts.SyncConfig

	if syncVault {
		webdavStorage := storage.NewWebDAVStorage(c.config.WebDAV)
		if err := webdavStorage.Connect(); err != nil {
			return err
		}
		if !webdavStorage.Exists() {
			return ErrRemoteVaultNotFound
		}
		data, err := webdavStorage.Read()
		if err != nil {
			return err
		}
		if err := c.storage.Write(data); err != nil {
			return err
		}
	}

	if syncConfig && c.config.WebDAV.ConfigRemotePath != "" && c.configPath != "" {
		configStorage := storage.NewWebDAVStorage(&types.WebDAVConfig{
			URL:                c.config.WebDAV.URL,
			Username:           c.config.WebDAV.Username,
			Password:           c.config.WebDAV.Password,
			RemotePath:         c.config.WebDAV.ConfigRemotePath,
			InsecureSkipVerify: c.config.WebDAV.InsecureSkipVerify,
		})
		if err := configStorage.Connect(); err != nil {
			return err
		}
		if !configStorage.Exists() {
			return ErrRemoteConfigNotFound
		}
		data, err := configStorage.Read()
		if err != nil {
			return err
		}
		if err := os.WriteFile(c.configPath, data, 0600); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) NewWebDAVStorage(cfg *types.WebDAVConfig) *storage.WebDAVStorage {
	return storage.NewWebDAVStorage(cfg)
}

var (
	ErrWebDAVNotConfigured  = storage.ErrStorageConnection
	ErrRemoteVaultNotFound  = storage.ErrStorageNotFound
	ErrRemoteConfigNotFound = storage.ErrStorageNotFound
)

func Encrypt(masterPassword string, salt []byte, plaintext string) (string, error) {
	cr := crypto.NewCrypto(masterPassword, salt)
	return cr.EncryptString(plaintext)
}

func Decrypt(masterPassword string, salt []byte, ciphertext string) (string, error) {
	cr := crypto.NewCrypto(masterPassword, salt)
	return cr.DecryptString(ciphertext)
}

func GenerateSalt() []byte {
	return crypto.GenerateSalt()
}
