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

// DeleteHandler 处理 DELETE FROM 语句，将其转换为 ES _delete_by_query 请求。
type DeleteHandler struct {
	client *elasticsearch.Client
}

// NewDeleteHandler 创建一个新的 DeleteHandler 实例。
func NewDeleteHandler(client *elasticsearch.Client) *DeleteHandler {
	return &DeleteHandler{client: client}
}

// reDeleteFrom 匹配 DELETE FROM 语句。
// 捕获组：1=表名, 2=WHERE 子句
var reDeleteFrom = regexp.MustCompile(
	`(?i)^\s*DELETE\s+FROM\s+` +
		`(\S+)\s+` + // 表名
		`WHERE\s+(.*?)` + // WHERE 子句
		`\s*;?\s*$`,
)

// Execute 解析 DELETE FROM SQL 并调用 ES POST /{index}/_delete_by_query 删除文档。
//
// 支持的语法：
//
//	DELETE FROM table_name WHERE field1=val1 AND field2=val2
//
// WHERE 子句中的值支持以下类型：
//   - 字符串：用单引号包裹，如 'hello'
//   - 数字：整数或浮点数，如 42, 3.14
//   - 布尔值：true / false（不区分大小写）
//   - 空值：null / NULL
func (h *DeleteHandler) Execute(ctx context.Context, sql string) (*types.Result, error) {
	tableName, whereConditions, err := parseDeleteFrom(sql)
	if err != nil {
		return nil, err
	}

	// 构造 ES _delete_by_query 请求体
	body, err := buildDeleteByQueryBody(whereConditions)
	if err != nil {
		return nil, err
	}

	// 调用 ES API 执行删除
	path := fmt.Sprintf("/%s/_delete_by_query", tableName)
	respBody, err := es.DoRequest(ctx, h.client, "POST", path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 解析 ES _delete_by_query 响应中的统计信息
	var esResp struct {
		Took     int   `json:"took"`
		Deleted  int   `json:"deleted"`
		Total    int   `json:"total"`
		Failures []any `json:"failures"`
	}
	if err := json.Unmarshal(respBody, &esResp); err != nil {
		return nil, fmt.Errorf("解析 ES 响应失败: %w", err)
	}

	return &types.Result{
		Meta: types.Meta{
			Status:   200,
			Message:  "删除完成",
			Endpoint: fmt.Sprintf("POST /%s/_delete_by_query", tableName),
			Stat: map[string]any{
				"删除行数":   esResp.Deleted,
				"匹配总数":   esResp.Total,
				"耗时(ms)": esResp.Took,
				"失败数":    len(esResp.Failures),
			},
		},
		Columns: []string{"结果"},
		Rows:    [][]any{{"删除完成"}},
		Source:  string(respBody),
	}, nil
}

// parseDeleteFrom 解析 DELETE FROM SQL 语句。
// 返回表名和 WHERE 条件列表。
func parseDeleteFrom(sql string) (string, map[string]any, error) {
	matches := reDeleteFrom.FindStringSubmatch(sql)
	if matches == nil {
		return "", nil, fmt.Errorf("SQL 语法错误：无法解析 DELETE FROM 语句。\n期望格式：DELETE FROM <表名> WHERE <条件列名>=<条件值> AND ...")
	}

	tableName := strings.TrimSpace(matches[1])
	wherePart := strings.TrimSpace(matches[2])

	if tableName == "" {
		return "", nil, fmt.Errorf("SQL 语法错误：缺少表名")
	}

	// 复用 update_handler.go 中的 parseWhereClause 解析 WHERE 子句
	whereConditions, err := parseWhereClause(wherePart)
	if err != nil {
		return "", nil, err
	}

	return tableName, whereConditions, nil
}

// buildDeleteByQueryBody 构造 ES _delete_by_query 的 JSON 请求体。
//
// 生成格式：
//
//	{
//	  "query": {
//	    "bool": {
//	      "must": [
//	        {"term": {"field": "value"}}
//	      ]
//	    }
//	  }
//	}
func buildDeleteByQueryBody(whereConditions map[string]any) (string, error) {
	mustClauses := make([]map[string]any, 0, len(whereConditions))
	for field, value := range whereConditions {
		mustClauses = append(mustClauses, map[string]any{
			"term": map[string]any{
				field: value,
			},
		})
	}

	body := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": mustClauses,
			},
		},
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("构造 JSON 请求体失败: %w", err)
	}

	return string(jsonBytes), nil
}
