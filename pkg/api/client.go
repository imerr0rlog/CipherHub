package api

import (
	"github.com/imerr0rlog/CipherHub/internal/crypto"
	"github.com/imerr0rlog/CipherHub/internal/storage"
	"github.com/imerr0rlog/CipherHub/internal/vault"
	"github.com/imerr0rlog/CipherHub/pkg/types"
)

type Client struct {
	manager *vault.Manager
	storage storage.Storage
	config  *types.Config
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

func (c *Client) InitVault(masterPassword string) error {
	return c.manager.Init(masterPassword)
}

func (c *Client) OpenVault(masterPassword string) error {
	return c.manager.Open(masterPassword)
}

func (c *Client) CloseVault() {
	c.manager.Close()
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

func (c *Client) Sync(remote storage.Storage) error {
	return c.manager.Sync(remote)
}

func (c *Client) Pull(remote storage.Storage, masterPassword string) error {
	return c.manager.Pull(remote, masterPassword)
}

func Encrypt(masterPassword string, salt []byte, plaintext string) (string, error) {
	c := crypto.NewCrypto(masterPassword, salt)
	return c.EncryptString(plaintext)
}

func Decrypt(masterPassword string, salt []byte, ciphertext string) (string, error) {
	c := crypto.NewCrypto(masterPassword, salt)
	return c.DecryptString(ciphertext)
}

func GenerateSalt() []byte {
	return crypto.GenerateSalt()
}
