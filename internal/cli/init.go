// Package cli 提供 CipherHub 的命令行界面实现
//
// 该包包含所有命令行命令的定义和实现，包括初始化密码库、添加/获取/删除条目、
// 配置管理、同步等功能。
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/imerr0rlog/CipherHub/internal/storage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

	// 尝试使用 term.ReadPassword 隐藏输入（终端模式）
	if term.IsTerminal(int(os.Stdin.Fd())) {
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // ReadPassword 不会输出换行
		return string(password), err
	}

	// 非 terminal 模式（如管道输入），回退到普通读取
	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func promptInput(prompt string) (string, error) {
	fmt.Print(prompt)

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
