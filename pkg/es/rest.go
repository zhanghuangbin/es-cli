package es

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
)

// DoRequest 使用 ES 客户端发送通用 HTTP 请求。
// 通过 client.Transport.Perform 发送请求，由 ES SDK 处理认证、TLS、重试等底层细节。
//
// 参数:
//   ctx    - 请求上下文
//   client - ES 客户端
//   method - HTTP 方法（GET, POST, PUT, DELETE 等）
//   path   - ES API 路径（如 /_sql, /my-index 等）
//   body   - 请求体，可为 nil
//
// 返回:
//   响应体字节数据和错误信息
func DoRequest(ctx context.Context, client *elasticsearch.Client, method, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("构造 HTTP 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Transport.Perform(req)
	if err != nil {
		return nil, fmt.Errorf("ES 请求失败 [%s %s]: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ES 返回错误 [%s %s]: 状态码 %d, 响应: %s", method, path, resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 ES 响应失败: %w", err)
	}

	return bodyBytes, nil
}
