package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cipherhub/cli/internal/storage"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new password vault",
	Long: `Initialize a new encrypted password vault.

You will be prompted to create a master password that will be used
to encrypt all your credentials. Make sure to remember this password
as it cannot be recovered.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if storage.NewLocalStorage(cfg.VaultPath).Exists() {
			return fmt.Errorf("vault already exists at %s", cfg.VaultPath)
		}

		fmt.Println("Creating a new CipherHub vault...")
		fmt.Println()

		password, err := promptPassword("Enter master password: ")
		if err != nil {
			return err
		}

		if len(password) < 8 {
			return fmt.Errorf("master password must be at least 8 characters")
		}

		confirm, err := promptPassword("Confirm master password: ")
		if err != nil {
			return err
		}

		if password != confirm {
			return fmt.Errorf("passwords do not match")
		}

		mgr, err := getVaultManager()
		if err != nil {
			return err
		}

		if err := mgr.Init(password); err != nil {
			return fmt.Errorf("failed to initialize vault: %w", err)
		}

		fmt.Println()
		fmt.Printf("✓ Vault created successfully at %s\n", cfg.VaultPath)
		fmt.Println()
		fmt.Println("You can now add entries with: cipherhub add <name>")

		return nil
	},
}

func promptPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(password), nil
}

func promptInput(prompt string) (string, error) {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}
