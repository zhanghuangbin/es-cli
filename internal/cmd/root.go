package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhanghuangbin/es-cli/internal/executor"
	"github.com/zhanghuangbin/es-cli/internal/formatter"
	"github.com/zhanghuangbin/es-cli/internal/repl"
	"github.com/zhanghuangbin/es-cli/internal/translator"
	"github.com/zhanghuangbin/es-cli/pkg/es"
)

var (
	addresses    []string
	username     string
	password     string
	passwordStdin bool
	caCert       string
)

var rootCmd = &cobra.Command{
	Use:   "es-cli",
	Short: "基于 SQL 的 Elasticsearch CLI",
	Long:  "一个交互式 REPL 工具，让你通过 SQL 语法查询和管理 Elasticsearch。",
	RunE: func(cmd *cobra.Command, args []string) error {
		if passwordStdin {
			if password != "" {
				return fmt.Errorf("错误: --password 和 --password-stdin 不能同时使用")
			}
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				password = strings.TrimSpace(scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("读取密码失败: %w", err)
			}
		}

		client, err := es.NewClient(es.Config{
			Addresses: addresses,
			Username:  username,
			Password:  password,
			CACert:    caCert,
		})
		if err != nil {
			return fmt.Errorf("连接 ES 失败: %w", err)
		}

		res, err := client.Ping()
		if err != nil {
			return fmt.Errorf("ping ES 失败: %w", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("ES ping 失败: %s", res.String())
		}

		fmt.Printf("已连接到 Elasticsearch (%s)\n", addresses)

		trans := translator.NewBuiltinTranslator(client)
		fmtr, _ := formatter.New("table")
		exec := executor.New(trans, fmtr, os.Stdout, client)

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
