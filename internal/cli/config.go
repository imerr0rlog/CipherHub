package cli

import (
	"encoding/json"
	"fmt"

	"github.com/cipherhub/cli/internal/storage"
	"github.com/spf13/cobra"
)

var (
	configWebDAVURL      string
	configWebDAVUser     string
	configWebDAVPassword string
	configWebDAVPath     string
	configSetLocal       bool
	configShow           bool
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
			cfg.DefaultStorage = types.StorageTypeWebDAV
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
			fmt.Printf("✓ WebDAV remote path set to %s\n", configWebDAVPath)
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
			fmt.Println("  --webdav-url URL       Set WebDAV server URL")
			fmt.Println("  --webdav-user USER     Set WebDAV username")
			fmt.Println("  --webdav-pass PASS     Set WebDAV password")
			fmt.Println("  --webdav-path PATH     Set remote vault path")
			fmt.Println("  --local                Set local as default storage")
			fmt.Println("  --show                 Show current configuration")
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
	configCmd.Flags().BoolVar(&configSetLocal, "local", false, "set local as default storage")
	configCmd.Flags().BoolVarP(&configShow, "show", "s", false, "show current configuration")
}
