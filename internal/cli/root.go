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
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(versionCmd)
}

func getVaultManager() (*vault.Manager, error) {
	if vaultMgr != nil && vaultMgr.IsOpen() {
		return vaultMgr, nil
	}

	st, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, err
	}

	vaultMgr = vault.NewManager(st)
	return vaultMgr, nil
}

func openVault(masterPassword string) (*vault.Manager, error) {
	mgr, err := getVaultManager()
	if err != nil {
		return nil, err
	}

	if mgr.IsOpen() {
		return mgr, nil
	}

	if err := mgr.Open(masterPassword); err != nil {
		return nil, err
	}

	return mgr, nil
}
