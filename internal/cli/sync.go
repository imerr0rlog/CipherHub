package cli

import (
	"fmt"

	"github.com/cipherhub/cli/internal/storage"
	"github.com/spf13/cobra"
)

var (
	syncPush  bool
	syncPull  bool
	syncForce bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync vault with WebDAV cloud storage",
	Long: `Sync vault with WebDAV cloud storage.

By default, sync pushes local changes to the remote storage.
Use --pull to download changes from remote.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.WebDAV == nil || cfg.WebDAV.URL == "" {
			return fmt.Errorf("WebDAV not configured. Run 'cipherhub config --webdav-url <url>' first")
		}

		masterPassword, err := promptPassword("Enter master password: ")
		if err != nil {
			return err
		}

		localStorage := storage.NewLocalStorage(cfg.VaultPath)
		mgr, err := openVault(masterPassword)
		if err != nil {
			return fmt.Errorf("failed to open vault: %w", err)
		}
		defer mgr.Close()

		webdavStorage := storage.NewWebDAVStorage(cfg.WebDAV)

		if err := webdavStorage.Connect(); err != nil {
			return fmt.Errorf("failed to connect to WebDAV: %w", err)
		}

		if syncPull {
			if !webdavStorage.Exists() {
				return fmt.Errorf("no remote vault found at %s", cfg.WebDAV.RemotePath)
			}

			if localStorage.Exists() && !syncForce {
				fmt.Print("This will overwrite local vault. Continue? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			if err := mgr.Pull(webdavStorage, masterPassword); err != nil {
				return fmt.Errorf("failed to pull from remote: %w", err)
			}

			fmt.Println("✓ Vault pulled from WebDAV successfully")
			return nil
		}

		if err := mgr.Sync(webdavStorage); err != nil {
			return fmt.Errorf("failed to sync to WebDAV: %w", err)
		}

		fmt.Println("✓ Vault synced to WebDAV successfully")
		return nil
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncPush, "push", true, "push local vault to remote")
	syncCmd.Flags().BoolVar(&syncPull, "pull", false, "pull remote vault to local")
	syncCmd.Flags().BoolVarP(&syncForce, "force", "f", false, "force overwrite without confirmation")
}
