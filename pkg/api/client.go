// Package api 提供了 CipherHub 密码库的公共 API 接口。
//
// 该包封装了密码库的核心功能，包括密码条目的增删改查、密码库的初始化和打开、
// WebDAV 同步等功能。使用此包可以方便地在应用程序中集成密码管理功能。
//
// 基本使用示例：
//
//	client, err := api.NewClient(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 初始化新密码库
//	err = client.InitVault("master-password")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 添加密码条目
//	entry, err := client.AddEntry("github", "user", "pass", "https://github.com", "", []string{"git"})
//	if err != nil {
//		log.Fatal(err)
//	}
package api

import (
	"encoding/json"
	"os"

	"github.com/imerr0rlog/CipherHub/internal/crypto"
	"github.com/imerr0rlog/CipherHub/internal/storage"
	"github.com/imerr0rlog/CipherHub/internal/vault"
	"github.com/imerr0rlog/CipherHub/pkg/types"
)

// Client 是 CipherHub 的主要客户端，提供所有密码库操作的公共接口。
//
// Client 封装了密码库管理器、存储后端和配置，提供了统一的 API 来操作密码条目、
// 管理密码库生命周期以及进行同步操作。
type Client struct {
	manager    *vault.Manager
	storage    storage.Storage
	config     *types.Config
	configPath string
}

// ClientOptions 用于配置 NewClientWithOptions 函数的选项。
//
// VaultPath 指定密码库文件的路径，ConfigPath 指定配置文件的路径。
// 如果未指定，将使用默认值。
type ClientOptions struct {
	VaultPath  string
	ConfigPath string
}

// NewClient 使用给定的配置创建一个新的 Client 实例。
//
// cfg 参数包含密码库的配置信息，包括密码库路径和其他设置。
// 返回初始化后的 Client 实例，或者在初始化失败时返回错误。
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

// NewClientWithOptions 使用选项创建一个新的 Client 实例。
//
// 此函数会尝试从配置文件加载配置，如果配置文件不存在则使用默认配置。
// opts 参数可以指定密码库路径和配置文件路径。
// 返回初始化后的 Client 实例，或者在初始化失败时返回错误。
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

// LoadConfig 从指定路径加载配置文件。
//
// path 参数是配置文件的路径。
// 返回加载的配置，或者在加载失败时返回错误。
func LoadConfig(path string) (*types.Config, error) {
	return storage.LoadConfig(path)
}

// SaveConfig 将配置保存到指定路径。
//
// path 参数是配置文件的路径，cfg 参数是要保存的配置。
// 返回保存成功时为 nil，否则返回错误。
func SaveConfig(path string, cfg *types.Config) error {
	return storage.SaveConfig(path, cfg)
}

// DefaultConfig 返回默认配置。
//
// 默认配置包含密码库的默认路径和其他默认设置。
func DefaultConfig() *types.Config {
	return types.DefaultConfig()
}

// Config 返回当前客户端使用的配置。
func (c *Client) Config() *types.Config {
	return c.config
}

// ConfigPath 返回当前配置文件的路径。
func (c *Client) ConfigPath() string {
	return c.configPath
}

// SetVaultPath 设置密码库的路径。
//
// path 参数是新的密码库文件路径。
// 设置后会重新初始化存储和密码库管理器。
func (c *Client) SetVaultPath(path string) {
	c.config.VaultPath = path
	c.storage = storage.NewLocalStorage(path)
	c.manager = vault.NewManager(c.storage)
}

// InitVault 使用主密码初始化一个新的密码库。
//
// masterPassword 参数是用于加密密码库的主密码。
// 返回初始化成功时为 nil，否则返回错误。
func (c *Client) InitVault(masterPassword string) error {
	return c.manager.Init(masterPassword)
}

// OpenVault 使用主密码打开已存在的密码库。
//
// masterPassword 参数是用于解密密码库的主密码。
// 返回打开成功时为 nil，否则返回错误。
func (c *Client) OpenVault(masterPassword string) error {
	return c.manager.Open(masterPassword)
}

// CloseVault 关闭当前打开的密码库。
//
// 关闭后，密码库将不再可访问，直到再次调用 OpenVault。
func (c *Client) CloseVault() {
	c.manager.Close()
}

// IsVaultOpen 检查密码库是否已打开。
//
// 返回 true 表示密码库已打开，false 表示未打开。
func (c *Client) IsVaultOpen() bool {
	return c.manager.IsOpen()
}

// VaultExists 检查密码库文件是否存在。
//
// 返回 true 表示密码库文件存在，false 表示不存在。
func (c *Client) VaultExists() bool {
	return c.storage.Exists()
}

// AddEntry 向密码库添加一个新的密码条目。
//
// name 参数是条目的名称（唯一标识），username 是用户名，password 是密码，
// url 是相关网站地址，notes 是备注信息，tags 是标签列表。
// 返回添加后的条目，或者在添加失败时返回错误。
func (c *Client) AddEntry(name, username, password, url, notes string, tags []string) (*types.Entry, error) {
	return c.manager.AddEntry(name, username, password, url, notes, tags)
}

// GetEntry 从密码库获取指定名称的条目。
//
// name 参数是要获取的条目的名称。
// 返回找到的条目，或者在未找到或获取失败时返回错误。
func (c *Client) GetEntry(name string) (*types.Entry, error) {
	return c.manager.GetEntry(name)
}

// GetDecryptedPassword 获取指定条目的解密后的密码。
//
// name 参数是要获取密码的条目的名称。
// 返回解密后的密码，或者在获取失败时返回错误。
func (c *Client) GetDecryptedPassword(name string) (string, error) {
	return c.manager.GetDecryptedPassword(name)
}

// GetDecryptedNotes 获取指定条目的解密后的备注。
//
// name 参数是要获取备注的条目的名称。
// 返回解密后的备注，或者在获取失败时返回错误。
func (c *Client) GetDecryptedNotes(name string) (string, error) {
	return c.manager.GetDecryptedNotes(name)
}

// ListEntries 列出密码库中的所有条目。
//
// 返回所有条目的列表，或者在获取失败时返回错误。
func (c *Client) ListEntries() ([]*types.Entry, error) {
	return c.manager.ListEntries()
}

// SearchEntries 在密码库中搜索匹配的条目。
//
// query 参数是搜索关键词，会在条目名称、用户名、URL、标签中进行匹配。
// 返回匹配的条目列表，或者在搜索失败时返回错误。
func (c *Client) SearchEntries(query string) ([]*types.Entry, error) {
	return c.manager.SearchEntries(query)
}

// UpdateEntry 更新密码库中指定条目的信息。
//
// name 参数是要更新的条目的名称，updates 参数是要更新的字段和值的映射。
// 返回更新后的条目，或者在更新失败时返回错误。
func (c *Client) UpdateEntry(name string, updates map[string]string) (*types.Entry, error) {
	return c.manager.UpdateEntry(name, updates)
}

// DeleteEntry 从密码库中删除指定名称的条目。
//
// name 参数是要删除的条目的名称。
// 返回删除成功时为 nil，否则返回错误。
func (c *Client) DeleteEntry(name string) error {
	return c.manager.DeleteEntry(name)
}

// GeneratePassword 生成一个指定长度的随机安全密码。
//
// length 参数是要生成的密码的长度。
// 返回生成的随机密码，或者在生成失败时返回错误。
func (c *Client) GeneratePassword(length int) (string, error) {
	return types.SecureRandomString(length)
}

// VaultInfo 返回密码库的信息。
//
// 返回一个包含密码库元数据的映射，如条目数量、创建时间等。
func (c *Client) VaultInfo() map[string]interface{} {
	return c.manager.VaultInfo()
}

// SyncOptions 用于配置同步操作的选项。
//
// SyncVault 控制是否同步密码库，SyncConfig 控制是否同步配置。
type SyncOptions struct {
	SyncVault  bool
	SyncConfig bool
}

// Sync 将本地密码库同步到远程存储。
//
// remote 参数是远程存储后端。
// 返回同步成功时为 nil，否则返回错误。
func (c *Client) Sync(remote storage.Storage) error {
	return c.manager.Sync(remote)
}

// Pull 从远程存储拉取密码库到本地。
//
// remote 参数是远程存储后端，masterPassword 是用于解密的主密码。
// 返回拉取成功时为 nil，否则返回错误。
func (c *Client) Pull(remote storage.Storage, masterPassword string) error {
	return c.manager.Pull(remote, masterPassword)
}

// SyncToWebDAV 将密码库和配置同步到 WebDAV 服务器。
//
// opts 参数控制同步哪些内容，默认为同步密码库和配置。
// 返回同步成功时为 nil，否则返回错误。
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

// PullFromWebDAV 从 WebDAV 服务器拉取密码库和配置。
//
// opts 参数控制拉取哪些内容，默认为拉取密码库和配置。
// 返回拉取成功时为 nil，否则返回错误。
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

// NewWebDAVStorage 创建一个新的 WebDAV 存储实例。
//
// cfg 参数是 WebDAV 配置。
// 返回初始化后的 WebDAV 存储实例。
func (c *Client) NewWebDAVStorage(cfg *types.WebDAVConfig) *storage.WebDAVStorage {
	return storage.NewWebDAVStorage(cfg)
}

var (
	// ErrWebDAVNotConfigured 表示 WebDAV 配置未设置或不完整的错误。
	ErrWebDAVNotConfigured  = storage.ErrStorageConnection
	// ErrRemoteVaultNotFound 表示远程密码库未找到的错误。
	ErrRemoteVaultNotFound  = storage.ErrStorageNotFound
	// ErrRemoteConfigNotFound 表示远程配置未找到的错误。
	ErrRemoteConfigNotFound = storage.ErrStorageNotFound
)

// Encrypt 使用主密码和盐值加密明文。
//
// masterPassword 参数是主密码，salt 参数是盐值，plaintext 参数是要加密的明文。
// 返回加密后的密文，或者在加密失败时返回错误。
func Encrypt(masterPassword string, salt []byte, plaintext string) (string, error) {
	cr := crypto.NewCrypto(masterPassword, salt)
	return cr.EncryptString(plaintext)
}

// Decrypt 使用主密码和盐值解密密文。
//
// masterPassword 参数是主密码，salt 参数是盐值，ciphertext 参数是要解密的密文。
// 返回解密后的明文，或者在解密失败时返回错误。
func Decrypt(masterPassword string, salt []byte, ciphertext string) (string, error) {
	cr := crypto.NewCrypto(masterPassword, salt)
	return cr.DecryptString(ciphertext)
}

// GenerateSalt 生成一个随机盐值。
//
// 返回生成的随机盐值字节数组。
func GenerateSalt() []byte {
	return crypto.GenerateSalt()
}
