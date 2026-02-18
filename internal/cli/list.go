// Package cli 提供 CipherHub 的命令行界面实现
//
// 该包包含所有命令行命令的定义和实现，包括初始化密码库、添加/获取/删除条目、
// 配置管理、同步等功能。
package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/imerr0rlog/CipherHub/pkg/types"
	"github.com/spf13/cobra"
)

var (
	listShowPasswords bool
	listSearch        string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all password entries",
	Long: `List all password entries in the vault.

Use --search to filter entries by name, username, or URL.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		masterPassword, err := promptPassword("Enter master password: ")
		if err != nil {
			return err
		}

		mgr, err := openVault(masterPassword)
		if err != nil {
			return fmt.Errorf("failed to open vault: %w", err)
		}
		defer mgr.Close()

		var entries []*types.Entry
		if listSearch != "" {
			entries, err = mgr.SearchEntries(listSearch)
			if err != nil {
				return err
			}
		} else {
			entries, err = mgr.ListEntries()
			if err != nil {
				return err
			}
		}

		if len(entries) == 0 {
			fmt.Println("No entries found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tUSERNAME\tURL\tUPDATED")

		for _, entry := range entries {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				entry.Name,
				entry.Username,
				entry.URL,
				entry.UpdatedAt.Format("2006-01-02"),
			)
		}

		w.Flush()
		fmt.Printf("\nTotal: %d entries\n", len(entries))

		return nil
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listShowPasswords, "passwords", "p", false, "show passwords (WARNING: insecure)")
	listCmd.Flags().StringVarP(&listSearch, "search", "s", "", "search entries by name, username, or URL")
}
