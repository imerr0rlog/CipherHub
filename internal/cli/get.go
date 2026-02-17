package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getShowPassword bool
	getShowNotes    bool
	getCopy         bool
)

var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Retrieve a password entry",
	Long: `Retrieve a password entry from the vault.

By default, only the entry details are shown without the password.
Use --password to display the password, or --copy to copy it to clipboard.`,
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

		entry, err := mgr.GetEntry(name)
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Printf("Name:     %s\n", entry.Name)
		fmt.Printf("Username: %s\n", entry.Username)
		fmt.Printf("URL:      %s\n", entry.URL)
		fmt.Printf("Created:  %s\n", entry.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("Updated:  %s\n", entry.UpdatedAt.Format("2006-01-02 15:04"))

		if len(entry.Tags) > 0 {
			fmt.Printf("Tags:     %v\n", entry.Tags)
		}

		if getShowPassword {
			password, err := mgr.GetDecryptedPassword(name)
			if err != nil {
				return fmt.Errorf("failed to decrypt password: %w", err)
			}
			fmt.Printf("Password: %s\n", password)
		}

		if getShowNotes && entry.Notes != "" {
			notes, err := mgr.GetDecryptedNotes(name)
			if err != nil {
				return fmt.Errorf("failed to decrypt notes: %w", err)
			}
			fmt.Printf("Notes:    %s\n", notes)
		}

		if getCopy {
			password, err := mgr.GetDecryptedPassword(name)
			if err != nil {
				return fmt.Errorf("failed to decrypt password: %w", err)
			}
			if err := copyToClipboard(password); err != nil {
				fmt.Printf("⚠ Failed to copy to clipboard: %v\n", err)
			} else {
				fmt.Println("✓ Password copied to clipboard")
			}
		}

		return nil
	},
}

func init() {
	getCmd.Flags().BoolVarP(&getShowPassword, "password", "p", false, "show the password")
	getCmd.Flags().BoolVarP(&getShowNotes, "notes", "n", false, "show the notes")
	getCmd.Flags().BoolVarP(&getCopy, "copy", "c", false, "copy password to clipboard")
}

func copyToClipboard(text string) error {
	return fmt.Errorf("clipboard not supported in this environment")
}
