package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zhanghuangbin/es-cli/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		version.Print(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
