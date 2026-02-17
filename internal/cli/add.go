package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	addUsername string
	addPassword string
	addURL      string
	addNotes    string
	addTags     string
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new password entry",
	Long: `Add a new password entry to the vault.

You will be prompted for the master password and any fields not provided
via flags. The password will be encrypted before storage.`,
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

		if addUsername == "" {
			addUsername, err = promptInput("Username: ")
			if err != nil {
				return err
			}
		}

		if addPassword == "" {
			addPassword, err = promptPassword("Password: ")
			if err != nil {
				return err
			}
		}

		var tags []string
		if addTags != "" {
			tags = strings.Split(addTags, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		}

		entry, err := mgr.AddEntry(name, addUsername, addPassword, addURL, addNotes, tags)
		if err != nil {
			return fmt.Errorf("failed to add entry: %w", err)
		}

		fmt.Printf("✓ Entry '%s' added successfully (ID: %s)\n", entry.Name, entry.ID)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addUsername, "username", "u", "", "username for the entry")
	addCmd.Flags().StringVarP(&addPassword, "password", "p", "", "password for the entry (will prompt if not provided)")
	addCmd.Flags().StringVarP(&addURL, "url", "U", "", "URL for the entry")
	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "notes for the entry")
	addCmd.Flags().StringVarP(&addTags, "tags", "t", "", "comma-separated tags")
}
