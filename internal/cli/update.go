package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateUsername string
	updatePassword string
	updateURL      string
	updateNotes    string
)

var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update an existing password entry",
	Long: `Update an existing password entry in the vault.

You can update one or more fields: username, password, URL, notes.
The master password will be prompted for verification.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		masterPassword, err := promptPassword("Enter master password: ")
		if err != nil {
			return err
		}

		mgr, err := openVault(masterPassword)
		if err != nil {
			return fmt.Errorf("failed to open vault: %w", err)
		}
		defer mgr.Close()

		updates := make(map[string]string)
		if updateUsername != "" {
			updates["username"] = updateUsername
		}
		if updatePassword != "" {
			updates["password"] = updatePassword
		} else {
			if updatePassword, err = promptPassword("New password (leave blank to keep existing): "); err != nil {
				return err
			}
			if updatePassword != "" {
				updates["password"] = updatePassword
			}
		}
		if updateURL != "" {
			updates["url"] = updateURL
		}
		if updateNotes != "" {
			updates["notes"] = updateNotes
		}

		if len(updates) == 0 {
			return fmt.Errorf("no updates specified")
		}

		entry, err := mgr.UpdateEntry(name, updates)
		if err != nil {
			return fmt.Errorf("failed to update entry: %w", err)
		}

		fmt.Printf("✓ Entry '%s' updated successfully\n", entry.Name)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateUsername, "username", "u", "", "new username")
	updateCmd.Flags().StringVarP(&updatePassword, "password", "p", "", "new password (leave blank to keep existing)")
	updateCmd.Flags().StringVarP(&updateURL, "url", "U", "", "new URL")
	updateCmd.Flags().StringVarP(&updateNotes, "notes", "n", "", "new notes")
}