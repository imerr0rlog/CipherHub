// Package cli 提供 CipherHub 的命令行界面实现
//
// 该包包含所有命令行命令的定义和实现，包括初始化密码库、添加/获取/删除条目、
// 配置管理、同步等功能。
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/imerr0rlog/CipherHub/internal/storage"
	"github.com/imerr0rlog/CipherHub/internal/vault"
	"github.com/imerr0rlog/CipherHub/pkg/types"
	"github.com/spf13/cobra"
)

var (
	syncPull       bool
	syncForce      bool
	syncVaultOnly  bool
	syncConfigOnly bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync vault and config with WebDAV cloud storage",
	Long: `Sync vault and config with WebDAV cloud storage.

By default, sync pushes both vault.json and config.json to remote.
Use --pull to download from remote.
Use --vault-only or --config-only to sync a single file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.WebDAV == nil || cfg.WebDAV.URL == "" {
			return fmt.Errorf("WebDAV not configured. Run 'cipherhub config --webdav-url <url>' first")
		}

		if syncVaultOnly && syncConfigOnly {
			return fmt.Errorf("cannot use --vault-only and --config-only together")
		}

		webdavStorage := storage.NewWebDAVStorage(cfg.WebDAV)
		if err := webdavStorage.Connect(); err != nil {
			return fmt.Errorf("failed to connect to WebDAV: %w", err)
		}

		if syncPull {
			return doPull(webdavStorage)
		}
		return doPush(webdavStorage)
	},
}

func doPush(webdavStorage *storage.WebDAVStorage) error {
	syncVault := !syncConfigOnly
	syncConfig := !syncVaultOnly

	if syncVault {
		if err := pushVault(webdavStorage); err != nil {
			return err
		}
	}

	if syncConfig {
		if err := pushConfig(webdavStorage); err != nil {
			return err
		}
	}

	return nil
}

func pushVault(webdavStorage *storage.WebDAVStorage) error {
	localStorage := storage.NewLocalStorage(cfg.VaultPath)
	if !localStorage.Exists() {
		return fmt.Errorf("local vault not found at %s. Run 'cipherhub init' first", cfg.VaultPath)
	}

	masterPassword, err := promptPassword("Enter master password: ")
	if err != nil {
		return err
	}

	mgr := vault.NewManager(localStorage)
	if err := mgr.Open(masterPassword); err != nil {
		return fmt.Errorf("failed to open vault: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Sync(webdavStorage); err != nil {
		return fmt.Errorf("failed to sync vault: %w", err)
	}

	fmt.Println("✓ Vault pushed to WebDAV")
	return nil
}

func pushConfig(webdavStorage *storage.WebDAVStorage) error {
	if cfg.WebDAV.ConfigRemotePath == "" {
		fmt.Println("⚠ Config remote path not set, skipping config sync")
		fmt.Println("  Run: cipherhub config --webdav-config-path /path/config.json")
		return nil
	}

	configData, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	configStorage := storage.NewWebDAVStorage(&types.WebDAVConfig{
		URL:                cfg.WebDAV.URL,
		Username:           cfg.WebDAV.Username,
		Password:           cfg.WebDAV.Password,
		RemotePath:         cfg.WebDAV.ConfigRemotePath,
		InsecureSkipVerify: cfg.WebDAV.InsecureSkipVerify,
	})
	if err := configStorage.Connect(); err != nil {
		return fmt.Errorf("failed to connect to WebDAV for config: %w", err)
	}

	if err := configStorage.Write(configData); err != nil {
		return fmt.Errorf("failed to sync config: %w", err)
	}

	fmt.Println("✓ Config pushed to WebDAV")
	return nil
}

func doPull(webdavStorage *storage.WebDAVStorage) error {
	syncVault := !syncConfigOnly
	syncConfig := !syncVaultOnly

	if !syncForce {
		targets := []string{}
		if syncVault {
			targets = append(targets, "vault")
		}
		if syncConfig {
			targets = append(targets, "config")
		}
		if len(targets) == 1 {
			fmt.Printf("This will overwrite local %s. Continue? [y/N]: ", targets[0])
		} else {
			fmt.Printf("This will overwrite local %s and %s. Continue? [y/N]: ", targets[0], targets[1])
		}
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if syncVault {
		if err := pullVault(webdavStorage); err != nil {
			return err
		}
	}

	if syncConfig {
		if err := pullConfig(); err != nil {
			return err
		}
	}

	return nil
}

func pullVault(webdavStorage *storage.WebDAVStorage) error {
	if !webdavStorage.Exists() {
		return fmt.Errorf("no remote vault found at %s", cfg.WebDAV.RemotePath)
	}

	data, err := webdavStorage.Read()
	if err != nil {
		return fmt.Errorf("failed to read remote vault: %w", err)
	}

	localStorage := storage.NewLocalStorage(cfg.VaultPath)
	if err := localStorage.Write(data); err != nil {
		return fmt.Errorf("failed to save local vault: %w", err)
	}

	fmt.Println("✓ Vault pulled from WebDAV")
	return nil
}

func pullConfig() error {
	if cfg.WebDAV.ConfigRemotePath == "" {
		fmt.Println("⚠ Config remote path not set, skipping config pull")
		return nil
	}

	configStorage := storage.NewWebDAVStorage(&types.WebDAVConfig{
		URL:                cfg.WebDAV.URL,
		Username:           cfg.WebDAV.Username,
		Password:           cfg.WebDAV.Password,
		RemotePath:         cfg.WebDAV.ConfigRemotePath,
		InsecureSkipVerify: cfg.WebDAV.InsecureSkipVerify,
	})
	if err := configStorage.Connect(); err != nil {
		return fmt.Errorf("failed to connect to WebDAV for config: %w", err)
	}

	if !configStorage.Exists() {
		return fmt.Errorf("no remote config found at %s", cfg.WebDAV.ConfigRemotePath)
	}

	data, err := configStorage.Read()
	if err != nil {
		return fmt.Errorf("failed to read remote config: %w", err)
	}

	var newCfg types.Config
	if err := json.Unmarshal(data, &newCfg); err != nil {
		return fmt.Errorf("invalid config format: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		return fmt.Errorf("failed to save local config: %w", err)
	}

	fmt.Println("✓ Config pulled from WebDAV")
	return nil
}

func init() {
	syncCmd.Flags().BoolVar(&syncPull, "pull", false, "pull from remote to local")
	syncCmd.Flags().BoolVarP(&syncForce, "force", "f", false, "force overwrite without confirmation")
	syncCmd.Flags().BoolVar(&syncVaultOnly, "vault-only", false, "sync only vault.json file")
	syncCmd.Flags().BoolVar(&syncConfigOnly, "config-only", false, "sync only config.json file")
}
