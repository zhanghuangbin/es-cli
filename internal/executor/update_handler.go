package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/translator"
	"github.com/zhanghuangbin/es-cli/pkg/es"
)

// UpdateHandler 处理 UPDATE 语句，将其转换为 ES _update_by_query 请求。
type UpdateHandler struct {
	client *elasticsearch.Client
}

// NewUpdateHandler 创建一个新的 UpdateHandler 实例。
func NewUpdateHandler(client *elasticsearch.Client) *UpdateHandler {
	return &UpdateHandler{client: client}
}

// reUpdate 匹配 UPDATE 语句。
// 捕获组：1=表名, 2=SET 子句, 3=WHERE 子句
var reUpdate = regexp.MustCompile(
	`(?i)^\s*UPDATE\s+` +
		`(\S+)\s+` + // 表名
		`SET\s+(.*?)` + // SET 子句
		`\s+WHERE\s+(.*?)` + // WHERE 子句
		`\s*;?\s*$`,
)

// Execute 解析 UPDATE SQL 并调用 ES POST /{index}/_update_by_query 更新文档。
//
// 支持的语法：
//
//	UPDATE table_name SET col1=val1, col2=val2 WHERE field1=val1 AND field2=val2
//
// SET 和 WHERE 子句中的值支持以下类型：
//   - 字符串：用单引号包裹，如 'hello'
//   - 数字：整数或浮点数，如 42, 3.14
//   - 布尔值：true / false（不区分大小写）
//   - 空值：null / NULL
func (h *UpdateHandler) Execute(ctx context.Context, sql string) (*translator.Result, error) {
	tableName, setFields, whereConditions, err := parseUpdate(sql)
	if err != nil {
		return nil, err
	}

	// 构造 ES _update_by_query 请求体
	body, err := buildUpdateByQueryBody(setFields, whereConditions)
	if err != nil {
		return nil, err
	}

	// 调用 ES API 执行更新
	path := fmt.Sprintf("/%s/_update_by_query", tableName)
	_, err = es.DoRequest(ctx, h.client, "POST", path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	return &translator.Result{
		Meta: translator.Meta{
			Status:  200,
			Message: "更新完成",
		},
		Columns: []string{"结果"},
		Rows:    [][]any{{"更新完成"}},
	}, nil
}

// parseUpdate 解析 UPDATE SQL 语句。
// 返回表名、SET 字段列表、WHERE 条件列表。
func parseUpdate(sql string) (string, map[string]any, map[string]any, error) {
	matches := reUpdate.FindStringSubmatch(sql)
	if matches == nil {
		return "", nil, nil, fmt.Errorf("SQL 语法错误：无法解析 UPDATE 语句。\n期望格式：UPDATE <表名> SET <列名>=<值>, ... WHERE <条件列名>=<条件值> AND ...")
	}

	tableName := strings.TrimSpace(matches[1])
	setPart := strings.TrimSpace(matches[2])
	wherePart := strings.TrimSpace(matches[3])

	if tableName == "" {
		return "", nil, nil, fmt.Errorf("SQL 语法错误：缺少表名")
	}

	// 解析 SET 子句
	setFields, err := parseSetClause(setPart)
	if err != nil {
		return "", nil, nil, err
	}

	// 解析 WHERE 子句
	whereConditions, err := parseWhereClause(wherePart)
	if err != nil {
		return "", nil, nil, err
	}

	return tableName, setFields, whereConditions, nil
}

// parseSetClause 解析 SET 子句，如 "name='hello', age=42"。
// 返回字段名到值的映射。
func parseSetClause(setPart string) (map[string]any, error) {
	if setPart == "" {
		return nil, fmt.Errorf("SQL 语法错误：SET 子句不能为空")
	}

	// 按逗号分割，但需要考虑单引号内的逗号
	pairs := splitValues(setPart)
	fields := make(map[string]any, len(pairs))

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		eqIdx := strings.Index(pair, "=")
		if eqIdx < 0 {
			return nil, fmt.Errorf("SQL 语法错误：SET 项 '%s' 格式不正确，期望格式：<列名>=<值>", pair)
		}

		key := strings.TrimSpace(pair[:eqIdx])
		rawVal := strings.TrimSpace(pair[eqIdx+1:])

		if key == "" {
			return nil, fmt.Errorf("SQL 语法错误：SET 子句中列名不能为空")
		}

		val, err := parseValue(rawVal)
		if err != nil {
			return nil, fmt.Errorf("SET 子句解析失败：%w", err)
		}

		fields[key] = val
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("SQL 语法错误：SET 子句至少需要指定一个字段")
	}

	return fields, nil
}

// reAndSplit 用于按 AND 关键字（不区分大小写）分割 WHERE 子句。
var reAndSplit = regexp.MustCompile(`(?i)\s+AND\s+`)

// parseWhereClause 解析 WHERE 子句，如 "name='hello' AND age=42"。
// 仅支持等值条件，条件之间用 AND 连接。
// 返回字段名到值的映射。
func parseWhereClause(wherePart string) (map[string]any, error) {
	if wherePart == "" {
		return nil, fmt.Errorf("SQL 语法错误：WHERE 子句不能为空")
	}

	// 按 AND 分割条件
	parts := reAndSplit.Split(wherePart, -1)
	conditions := make(map[string]any, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		eqIdx := strings.Index(part, "=")
		if eqIdx < 0 {
			return nil, fmt.Errorf("SQL 语法错误：WHERE 条件 '%s' 格式不正确，期望格式：<列名>=<值>", part)
		}

		key := strings.TrimSpace(part[:eqIdx])
		rawVal := strings.TrimSpace(part[eqIdx+1:])

		if key == "" {
			return nil, fmt.Errorf("SQL 语法错误：WHERE 子句中列名不能为空")
		}

		val, err := parseValue(rawVal)
		if err != nil {
			return nil, fmt.Errorf("WHERE 子句解析失败：%w", err)
		}

		conditions[key] = val
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("SQL 语法错误：WHERE 子句至少需要指定一个条件")
	}

	return conditions, nil
}

// buildUpdateByQueryBody 构造 ES _update_by_query 的 JSON 请求体。
//
// 生成格式：
//
//	{
//	  "script": {
//	    "source": "ctx._source.field1 = params.field1; ctx._source.field2 = params.field2",
//	    "params": { "field1": "value1", "field2": "value2" }
//	  },
//	  "query": {
//	    "bool": {
//	      "must": [
//	        {"term": {"field": "value"}}
//	      ]
//	    }
//	  }
//	}
func buildUpdateByQueryBody(setFields map[string]any, whereConditions map[string]any) (string, error) {
	// 构造 script
	var scriptParts []string
	params := make(map[string]any, len(setFields))
	for field, value := range setFields {
		scriptParts = append(scriptParts, fmt.Sprintf("ctx._source.%s = params.%s", field, field))
		params[field] = value
	}

	script := map[string]any{
		"source": strings.Join(scriptParts, "; "),
		"params": params,
	}

	// 构造 query
	mustClauses := make([]map[string]any, 0, len(whereConditions))
	for field, value := range whereConditions {
		mustClauses = append(mustClauses, map[string]any{
			"term": map[string]any{
				field: value,
			},
		})
	}

	query := map[string]any{
		"bool": map[string]any{
			"must": mustClauses,
		},
	}

	body := map[string]any{
		"script": script,
		"query":  query,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("构造 JSON 请求体失败: %w", err)
	}

	return string(jsonBytes), nil
}
