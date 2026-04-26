package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhanghuangbin/es-cli/internal/executor"
	replFmt "github.com/zhanghuangbin/es-cli/internal/formatter/repl"
	"github.com/zhanghuangbin/es-cli/internal/repl"
)

var (
	addresses     []string
	username      string
	password      string
	passwordStdin bool
	caCert        string
)

var rootCmd = &cobra.Command{
	Use:   "es-cli",
	Short: "基于 SQL 的 Elasticsearch CLI",
	Long:  "一个交互式 REPL 工具，让你通过 SQL 语法查询和管理 Elasticsearch。",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newESClient()
		if err != nil {
			return err
		}

		fmt.Printf("已连接到 Elasticsearch (%s)\n", addresses)

		fmtr, _ := replFmt.NewFormatter("table")
		exec := executor.New(fmtr, os.Stdout, client)

		r := repl.New(exec, client, addresses)
		r.Run()
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&addresses, "address", []string{"http://localhost:9200"}, "Elasticsearch 地址")
	rootCmd.PersistentFlags().StringVar(&username, "username", "", "Elasticsearch 用户名")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Elasticsearch 密码（建议使用 --password-stdin）")
	rootCmd.PersistentFlags().BoolVar(&passwordStdin, "password-stdin", false, "从标准输入读取密码（每行一个）")
	rootCmd.PersistentFlags().StringVar(&caCert, "ca-cert", "", "TLS CA 证书路径")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
