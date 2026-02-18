package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/imerr0rlog/CipherHub/pkg/types"
)

// LocalStorage 实现了基于本地文件系统的存储
//
// 该类型使用本地文件系统作为密码库的存储介质。
type LocalStorage struct {
	path string
}

// NewLocalStorage 创建一个新的本地存储实例
//
// path 为密码库文件的完整路径。
// 返回创建的 LocalStorage 实例。
func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{path: path}
}

// Read 从本地文件读取密码库数据
//
// 返回读取到的字节数据，如果文件不存在则返回 ErrStorageNotFound。
func (s *LocalStorage) Read() ([]byte, error) {
	if !s.Exists() {
		return nil, ErrStorageNotFound
	}
	return os.ReadFile(s.path)
}

// Write 将密码库数据写入本地文件
//
// data 为要写入的字节数据。
// 写入时会先创建临时文件，成功后再重命名，以避免写入过程中程序崩溃导致数据损坏。
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

// Exists 检查本地文件是否存在
//
// 存在返回 true，否则返回 false。
func (s *LocalStorage) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

// Delete 删除本地文件
//
// 如果文件不存在则返回 ErrStorageNotFound。
func (s *LocalStorage) Delete() error {
	if !s.Exists() {
		return ErrStorageNotFound
	}
	return os.Remove(s.path)
}

// Type 返回存储类型
//
// 返回 types.StorageTypeLocal。
func (s *LocalStorage) Type() types.StorageType {
	return types.StorageTypeLocal
}

// Path 返回本地文件路径
//
// 返回存储文件的完整路径。
func (s *LocalStorage) Path() string {
	return s.path
}

// LoadConfig 从指定路径加载配置文件
//
// path 为配置文件路径。
// 如果文件不存在，返回默认配置而不报错。
// 返回加载的配置对象，如果加载失败则返回错误。
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

// SaveConfig 将配置保存到指定路径
//
// path 为保存路径，cfg 为要保存的配置对象。
// 会自动创建父目录（权限 0700），配置文件权限为 0600。
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

// GetConfigPath 获取默认配置文件路径
//
// 默认配置文件位于可执行程序所在目录下的 config.json。
// 返回配置文件路径，如果获取可执行程序路径失败则返回错误。
func GetConfigPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "config.json"), nil
}

// LoadOrCreateConfig 加载或创建配置文件
//
// 使用默认配置文件路径。
// 返回配置对象、配置文件路径，如果加载失败则返回错误。
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

// LoadOrCreateConfigWithPath 从指定路径加载或创建配置文件
//
// customPath 为自定义配置文件路径，如果为空则使用默认路径。
// 返回配置对象、配置文件路径，如果加载失败则返回错误。
func LoadOrCreateConfigWithPath(customPath string) (*types.Config, string, error) {
	configPath := customPath
	if configPath == "" {
		var err error
		configPath, err = GetConfigPath()
		if err != nil {
			return nil, "", err
		}
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = types.DefaultConfig()
			if customPath != "" {
				cfg.VaultPath = filepath.Join(filepath.Dir(customPath), "vault.json")
			}
			return cfg, configPath, nil
		}
		return nil, "", err
	}

	return cfg, configPath, nil
}
