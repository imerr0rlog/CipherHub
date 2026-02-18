// Package cli 提供 CipherHub 的命令行界面实现
//
// 该包包含所有命令行命令的定义和实现，包括初始化密码库、添加/获取/删除条目、
// 配置管理、同步等功能。
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a password entry",
	Long: `Delete a password entry from the vault.

This action is irreversible. Use --force to skip confirmation.`,
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

		if !deleteForce {
			fmt.Printf("Are you sure you want to delete '%s' (%s)? [y/N]: ", entry.Name, entry.Username)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		if err := mgr.DeleteEntry(name); err != nil {
			return fmt.Errorf("failed to delete entry: %w", err)
		}

		fmt.Printf("✓ Entry '%s' deleted\n", name)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation")
}
