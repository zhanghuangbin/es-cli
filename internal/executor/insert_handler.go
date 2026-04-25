package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/translator"
	"github.com/zhanghuangbin/es-cli/pkg/es"
)

// InsertHandler 处理 INSERT INTO 语句，将其转换为 ES 文档索引请求。
type InsertHandler struct {
	client *elasticsearch.Client
}

// NewInsertHandler 创建一个新的 InsertHandler 实例。
func NewInsertHandler(client *elasticsearch.Client) *InsertHandler {
	return &InsertHandler{client: client}
}

// reInsertInto 匹配 INSERT INTO 语句。
// 捕获组：1=表名, 2=列名部分, 3=值部分
var reInsertInto = regexp.MustCompile(
	`(?i)^\s*INSERT\s+INTO\s+` +
		`(\S+)\s*` + // 表名
		`\(\s*(.*?)\s*\)` + // 列名列表
		`\s+VALUES\s*` +
		`\(\s*(.*?)\s*\)` + // 值列表
		`\s*;?\s*$`,
)

// Execute 解析 INSERT INTO SQL 并调用 ES POST /{index}/_doc 插入文档。
//
// 支持的语法：
//
//	INSERT INTO table_name (col1, col2, ...) VALUES (val1, val2, ...)
//
// 值支持以下类型：
//   - 字符串：用单引号包裹，如 'hello'
//   - 数字：整数或浮点数，如 42, 3.14
//   - 布尔值：true / false（不区分大小写）
//   - 空值：null / NULL
func (h *InsertHandler) Execute(ctx context.Context, sql string) (*translator.Result, error) {
	tableName, doc, err := parseInsertInto(sql)
	if err != nil {
		return nil, err
	}

	// 序列化文档为 JSON
	jsonBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("构造 JSON 文档失败: %w", err)
	}

	// 调用 ES API 插入文档
	path := fmt.Sprintf("/%s/_doc", tableName)
	_, err = es.DoRequest(ctx, h.client, "POST", path, strings.NewReader(string(jsonBytes)))
	if err != nil {
		return nil, err
	}

	return &translator.Result{
		Meta: translator.Meta{
			Status:  200,
			Message: "文档插入成功",
		},
		Columns: []string{"结果"},
		Rows:    [][]any{{"文档插入成功"}},
	}, nil
}

// parseInsertInto 解析 INSERT INTO SQL 语句。
// 返回表名和文档（列名到值的映射）。
func parseInsertInto(sql string) (string, map[string]any, error) {
	matches := reInsertInto.FindStringSubmatch(sql)
	if matches == nil {
		return "", nil, fmt.Errorf("SQL 语法错误：无法解析 INSERT INTO 语句。\n期望格式：INSERT INTO <表名> (<列名1>, <列名2>, ...) VALUES (<值1>, <值2>, ...)")
	}

	tableName := strings.TrimSpace(matches[1])
	columnsPart := strings.TrimSpace(matches[2])
	valuesPart := strings.TrimSpace(matches[3])

	if tableName == "" {
		return "", nil, fmt.Errorf("SQL 语法错误：缺少表名")
	}

	// 解析列名
	columns, err := parseInsertColumns(columnsPart)
	if err != nil {
		return "", nil, err
	}

	// 解析值
	values, err := parseInsertValues(valuesPart)
	if err != nil {
		return "", nil, err
	}

	// 列数和值数必须一致
	if len(columns) != len(values) {
		return "", nil, fmt.Errorf("SQL 语法错误：列数(%d)与值数(%d)不匹配", len(columns), len(values))
	}

	// 构造文档
	doc := make(map[string]any, len(columns))
	for i, col := range columns {
		doc[col] = values[i]
	}

	return tableName, doc, nil
}

// parseInsertColumns 解析 INSERT INTO 的列名部分，如 "name, age, active"。
func parseInsertColumns(columnsPart string) ([]string, error) {
	if columnsPart == "" {
		return nil, fmt.Errorf("SQL 语法错误：列名列表不能为空")
	}

	parts := strings.Split(columnsPart, ",")
	columns := make([]string, 0, len(parts))

	for _, part := range parts {
		col := strings.TrimSpace(part)
		if col == "" {
			continue
		}
		columns = append(columns, col)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("SQL 语法错误：至少需要指定一个列名")
	}

	return columns, nil
}

// parseInsertValues 解析 INSERT INTO 的值部分，如 "'hello', 42, true, null"。
// 支持字符串（单引号）、数字、布尔值、null。
func parseInsertValues(valuesPart string) ([]any, error) {
	if valuesPart == "" {
		return nil, fmt.Errorf("SQL 语法错误：值列表不能为空")
	}

	// 按逗号分割值，但需要考虑单引号内的逗号
	rawValues := splitValues(valuesPart)
	values := make([]any, 0, len(rawValues))

	for _, raw := range rawValues {
		val, err := parseValue(raw)
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("SQL 语法错误：至少需要指定一个值")
	}

	return values, nil
}

// splitValues 按逗号分割值列表，但忽略单引号内的逗号。
func splitValues(s string) []string {
	var result []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '\'' && !inQuote:
			inQuote = true
			current.WriteByte(ch)
		case ch == '\'' && inQuote:
			// 处理转义的单引号 ''
			if i+1 < len(s) && s[i+1] == '\'' {
				current.WriteByte('\'')
				current.WriteByte('\'')
				i++
			} else {
				inQuote = false
				current.WriteByte(ch)
			}
		case ch == ',' && !inQuote:
			result = append(result, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteByte(ch)
		}
	}

	// 添加最后一个值
	last := strings.TrimSpace(current.String())
	if last != "" {
		result = append(result, last)
	}

	return result
}

// parseValue 解析单个值，返回对应的 Go 类型。
func parseValue(raw string) (any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("SQL 语法错误：值不能为空")
	}

	// 字符串值：以单引号包裹
	if strings.HasPrefix(raw, "'") && strings.HasSuffix(raw, "'") && len(raw) >= 2 {
		// 去除首尾引号，并处理转义的单引号
		inner := raw[1 : len(raw)-1]
		inner = strings.ReplaceAll(inner, "''", "'")
		return inner, nil
	}

	// null / NULL
	if strings.EqualFold(raw, "null") {
		return nil, nil
	}

	// 布尔值
	if strings.EqualFold(raw, "true") {
		return true, nil
	}
	if strings.EqualFold(raw, "false") {
		return false, nil
	}

	// 整数
	if intVal, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return intVal, nil
	}

	// 浮点数
	if floatVal, err := strconv.ParseFloat(raw, 64); err == nil {
		return floatVal, nil
	}

	return nil, fmt.Errorf("SQL 语法错误：无法解析值 '%s'，支持的类型：字符串('...')、数字、布尔值(true/false)、空值(null)", raw)
}
