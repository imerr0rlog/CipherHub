// Package cli 提供 CipherHub 的命令行界面实现
//
// 该包包含所有命令行命令的定义和实现，包括初始化密码库、添加/获取/删除条目、
// 配置管理、同步等功能。
package cli

import (
	"encoding/json"
	"fmt"

	"github.com/imerr0rlog/CipherHub/internal/storage"
	"github.com/imerr0rlog/CipherHub/pkg/types"
	"github.com/spf13/cobra"
)

var (
	configWebDAVURL        string
	configWebDAVUser       string
	configWebDAVPassword   string
	configWebDAVPath       string
	configWebDAVConfigPath string
	configSetLocal         bool
	configShow             bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CipherHub configuration",
	Long: `Manage CipherHub configuration settings.

You can configure WebDAV for cloud sync, change the default storage,
or view current settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if configShow {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("Configuration file: %s\n\n", cfgPath)
			fmt.Println(string(data))
			return nil
		}

		changed := false

		if configSetLocal {
			cfg.DefaultStorage = types.StorageTypeLocal
			changed = true
			fmt.Println("✓ Default storage set to local")
		}

		if configWebDAVURL != "" {
			if cfg.WebDAV == nil {
				cfg.WebDAV = &types.WebDAVConfig{}
			}
			cfg.WebDAV.URL = configWebDAVURL
			changed = true
			fmt.Printf("✓ WebDAV URL set to %s\n", configWebDAVURL)
		}

		if configWebDAVUser != "" {
			if cfg.WebDAV == nil {
				cfg.WebDAV = &types.WebDAVConfig{}
			}
			cfg.WebDAV.Username = configWebDAVUser
			changed = true
			fmt.Printf("✓ WebDAV username set to %s\n", configWebDAVUser)
		}

		if configWebDAVPassword != "" {
			if cfg.WebDAV == nil {
				cfg.WebDAV = &types.WebDAVConfig{}
			}
			cfg.WebDAV.Password = configWebDAVPassword
			changed = true
			fmt.Println("✓ WebDAV password updated")
		}

		if configWebDAVPath != "" {
			if cfg.WebDAV == nil {
				cfg.WebDAV = &types.WebDAVConfig{}
			}
			cfg.WebDAV.RemotePath = configWebDAVPath
			changed = true
			fmt.Printf("✓ WebDAV vault path set to %s\n", configWebDAVPath)
		}

		if configWebDAVConfigPath != "" {
			if cfg.WebDAV == nil {
				cfg.WebDAV = &types.WebDAVConfig{}
			}
			cfg.WebDAV.ConfigRemotePath = configWebDAVConfigPath
			changed = true
			fmt.Printf("✓ WebDAV config path set to %s\n", configWebDAVConfigPath)
		}

		if !changed {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Printf("Configuration file: %s\n\n", cfgPath)
			fmt.Println(string(data))
			fmt.Println()
			fmt.Println("Use flags to modify configuration:")
			fmt.Println("  --webdav-url URL         Set WebDAV server URL")
			fmt.Println("  --webdav-user USER       Set WebDAV username")
			fmt.Println("  --webdav-pass PASS       Set WebDAV password")
			fmt.Println("  --webdav-path PATH       Set remote vault path")
			fmt.Println("  --webdav-config-path PATH Set remote config path")
			fmt.Println("  --local                  Set local as default storage")
			fmt.Println("  --show                   Show current configuration")
			return nil
		}

		if err := storage.SaveConfig(cfgPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("\nConfiguration saved to %s\n", cfgPath)
		return nil
	},
}

func init() {
	configCmd.Flags().StringVar(&configWebDAVURL, "webdav-url", "", "WebDAV server URL")
	configCmd.Flags().StringVar(&configWebDAVUser, "webdav-user", "", "WebDAV username")
	configCmd.Flags().StringVar(&configWebDAVPassword, "webdav-pass", "", "WebDAV password")
	configCmd.Flags().StringVar(&configWebDAVPath, "webdav-path", "", "remote vault path on WebDAV")
	configCmd.Flags().StringVar(&configWebDAVConfigPath, "webdav-config-path", "", "remote config path on WebDAV")
	configCmd.Flags().BoolVar(&configSetLocal, "local", false, "set local as default storage")
	configCmd.Flags().BoolVarP(&configShow, "show", "s", false, "show current configuration")
}
