package cli

import (
	"fmt"

	"github.com/imerr0rlog/CipherHub/pkg/types"
	"github.com/spf13/cobra"
)

var generateLength int

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a secure random password",
	Long: `Generate a secure random password.

The password is generated using cryptographically secure random bytes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		password := types.SecureRandomString(generateLength)
		fmt.Println(password)
		return nil
	},
}

func init() {
	generateCmd.Flags().IntVarP(&generateLength, "length", "l", 16, "password length")
}
