package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/imerr0rlog/CipherHub/pkg/types"
)

type LocalStorage struct {
	path string
}

func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{path: path}
}

func (s *LocalStorage) Read() ([]byte, error) {
	if !s.Exists() {
		return nil, ErrStorageNotFound
	}
	return os.ReadFile(s.path)
}

func (s *LocalStorage) Write(data []byte) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.path)
}

func (s *LocalStorage) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

func (s *LocalStorage) Delete() error {
	if !s.Exists() {
		return ErrStorageNotFound
	}
	return os.Remove(s.path)
}

func (s *LocalStorage) Type() types.StorageType {
	return types.StorageTypeLocal
}

func (s *LocalStorage) Path() string {
	return s.path
}

func LoadConfig(path string) (*types.Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return types.DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg types.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func SaveConfig(path string, cfg *types.Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cipherhub", "config.json"), nil
}

func LoadOrCreateConfig() (*types.Config, string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, "", err
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = types.DefaultConfig()
			return cfg, configPath, nil
		}
		return nil, "", err
	}

	return cfg, configPath, nil
}
