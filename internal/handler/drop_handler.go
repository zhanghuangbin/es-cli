package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/types"
	"github.com/zhanghuangbin/es-cli/pkg/es"
)

// DropHandler 处理 DROP TABLE 语句，将其转换为 ES 索引删除请求。
type DropHandler struct {
	client *elasticsearch.Client
}

// NewDropHandler 创建一个新的 DropHandler 实例。
func NewDropHandler(client *elasticsearch.Client) *DropHandler {
	return &DropHandler{client: client}
}

// reDropTable 匹配 DROP TABLE 语句。
// 捕获组：1=表名
var reDropTable = regexp.MustCompile(
	`(?i)^\s*DROP\s+TABLE\s+(\S+?)\s*;?\s*$`,
)

// Execute 解析 DROP TABLE SQL 并调用 ES DELETE /{index} 删除索引。
//
// 支持的语法：
//
//	DROP TABLE table_name
func (h *DropHandler) Execute(ctx context.Context, sql string) (*types.Result, error) {
	tableName, err := parseDropTable(sql)
	if err != nil {
		return nil, err
	}

	// 调用 ES API 删除索引
	path := fmt.Sprintf("/%s", tableName)
	respBody, err := es.DoRequest(ctx, h.client, "DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	// 解析 ES 响应中的确认信息
	var esResp struct {
		Acknowledged bool `json:"acknowledged"`
	}
	if err := json.Unmarshal(respBody, &esResp); err != nil {
		return nil, fmt.Errorf("解析 ES 响应失败: %w", err)
	}

	return &types.Result{
		Meta: types.Meta{
			Status:   200,
			Message:  fmt.Sprintf("索引 %s 删除成功", tableName),
			Endpoint: fmt.Sprintf("DELETE /%s", tableName),
			Stat: map[string]any{
				"索引名":          tableName,
				"acknowledged": esResp.Acknowledged,
			},
		},
		Columns: []string{"结果"},
		Rows:    [][]any{{fmt.Sprintf("索引 %s 删除成功", tableName)}},
		Source:  string(respBody),
	}, nil
}

// parseDropTable 解析 DROP TABLE SQL 语句。
// 返回表名。
func parseDropTable(sql string) (string, error) {
	matches := reDropTable.FindStringSubmatch(sql)
	if matches == nil {
		return "", fmt.Errorf("SQL 语法错误：无法解析 DROP TABLE 语句。\n期望格式：DROP TABLE <表名>")
	}

	tableName := strings.TrimSpace(matches[1])
	if tableName == "" {
		return "", fmt.Errorf("SQL 语法错误：缺少表名")
	}

	return tableName, nil
}
