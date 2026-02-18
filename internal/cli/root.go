// Package cli 提供 CipherHub 的命令行界面实现
//
// 该包包含所有命令行命令的定义和实现，包括初始化密码库、添加/获取/删除条目、
// 配置管理、同步等功能。
package cli

import (
	"fmt"
	"os"

	"github.com/imerr0rlog/CipherHub/internal/storage"
	"github.com/imerr0rlog/CipherHub/internal/vault"
	"github.com/imerr0rlog/CipherHub/pkg/types"
	"github.com/spf13/cobra"
)

var (
	cfg            *types.Config
	cfgPath        string
	vaultMgr       *vault.Manager
	flagConfigPath string
	flagVaultPath  string
)

var rootCmd = &cobra.Command{
	Use:   "cipherhub",
	Short: "A secure password manager for the command line",
	Long: `CipherHub is a secure, encrypted password manager that stores your
credentials locally or syncs them to WebDAV cloud storage.

All passwords are encrypted using AES-256-GCM with keys derived from
your master password using Argon2id.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, cfgPath, err = storage.LoadOrCreateConfigWithPath(flagConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if flagVaultPath != "" {
			cfg.VaultPath = flagVaultPath
		}
		return nil
	},
}

// Execute 运行 CipherHub 命令行应用程序
//
// 该函数解析命令行参数并执行相应的命令。如果执行过程中发生错误，
// 会将错误信息输出到标准错误流并以状态码 1 退出程序。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagConfigPath, "config", "", "config file path (default: ./config.json)")
	rootCmd.PersistentFlags().StringVar(&flagVaultPath, "vault", "", "vault file path (default: ./vault.json)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(versionCmd)
}

func getVaultManager() (*vault.Manager, error) {
	st, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, err
	}

	return vault.NewManager(st), nil
}

func openVault(masterPassword string) (*vault.Manager, error) {
	mgr, err := getVaultManager()
	if err != nil {
		return nil, err
	}

	if err := mgr.Open(masterPassword); err != nil {
		return nil, err
	}

	return mgr, nil
}
