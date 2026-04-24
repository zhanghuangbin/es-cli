package es

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewClient(cfg Config) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	if cfg.CACert != "" {
		caCert, err := os.ReadFile(cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("读取 CA 证书失败: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("解析 CA 证书失败")
		}
		esCfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 ES 客户端失败: %w", err)
	}
	return client, nil
}
