package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/pkg/es"
)

func newESClient() (*elasticsearch.Client, error) {
	if passwordStdin {
		if password != "" {
			return nil, fmt.Errorf("错误: --password 和 --password-stdin 不能同时使用")
		}
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			password = strings.TrimSpace(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("读取密码失败: %w", err)
		}
	}

	client, err := es.NewClient(es.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
		CACert:    caCert,
	})
	if err != nil {
		return nil, fmt.Errorf("连接 ES 失败: %w", err)
	}

	res, err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping ES 失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES ping 失败: %s", res.String())
	}

	return client, nil
}
