// Package cli 提供 CipherHub 的命令行界面实现
//
// 该包包含所有命令行命令的定义和实现，包括初始化密码库、添加/获取/删除条目、
// 配置管理、同步等功能。
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
		password, err := types.SecureRandomString(generateLength)
		if err != nil {
			return err
		}
		fmt.Println(password)
		return nil
	},
}

func init() {
	generateCmd.Flags().IntVarP(&generateLength, "length", "l", 16, "password length")
}
