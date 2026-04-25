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

// sqlTypeMapping 定义 SQL 数据类型到 ES 字段类型的映射。
var sqlTypeMapping = map[string]string{
	"INT":       "integer",
	"INTEGER":   "integer",
	"BIGINT":    "long",
	"LONG":      "long",
	"FLOAT":     "float",
	"DOUBLE":    "double",
	"BOOLEAN":   "boolean",
	"BOOL":      "boolean",
	"TEXT":      "text",
	"KEYWORD":   "keyword",
	"VARCHAR":   "keyword",
	"DATE":      "date",
	"DATETIME":  "date",
	"TIMESTAMP": "date",
	"OBJECT":    "object",
	"JSON":      "object",
	"NESTED":    "nested",
}

// columnDef 表示一个列定义（字段名 + ES 类型）。
type columnDef struct {
	Name string
	Type string
}

// CreateHandler 处理 CREATE TABLE 语句，将其转换为 ES 索引创建请求。
type CreateHandler struct {
	client *elasticsearch.Client
}

// NewCreateHandler 创建一个新的 CreateHandler 实例。
func NewCreateHandler(client *elasticsearch.Client) *CreateHandler {
	return &CreateHandler{client: client}
}

// Execute 解析 CREATE TABLE SQL 并调用 ES PUT /{index} 创建索引。
//
// 支持的语法：
//
//	CREATE TABLE table_name (col1 TYPE1, col2 TYPE2, ...) [SETTINGS (k1=v1, k2=v2, ...)]
func (h *CreateHandler) Execute(ctx context.Context, sql string) (*translator.Result, error) {
	tableName, columns, settings, err := parseCreateTable(sql)
	if err != nil {
		return nil, err
	}

	// 构造 ES 请求体
	body, err := buildCreateIndexBody(columns, settings)
	if err != nil {
		return nil, err
	}

	// 调用 ES API 创建索引
	path := fmt.Sprintf("/%s", tableName)
	_, err = es.DoRequest(ctx, h.client, "PUT", path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	return &translator.Result{
		Meta: translator.Meta{
			Status:  200,
			Message: fmt.Sprintf("索引 %s 创建成功", tableName),
		},
		Columns: []string{"结果"},
		Rows:    [][]any{{fmt.Sprintf("索引 %s 创建成功", tableName)}},
	}, nil
}

// reCreateTable 匹配 CREATE TABLE 语句的整体结构。
// 捕获组：1=表名, 2=列定义部分, 3=SETTINGS 子句（可选）
var reCreateTable = regexp.MustCompile(
	`(?i)^\s*CREATE\s+TABLE\s+` +
		`(\S+)\s*` + // 表名
		`\(\s*(.*?)\s*\)` + // 列定义
		`(?:\s+SETTINGS\s*\(\s*(.*?)\s*\))?\s*;?\s*$`, // 可选 SETTINGS
)

// parseCreateTable 解析 CREATE TABLE SQL 语句。
// 返回表名、列定义列表、settings 映射。
func parseCreateTable(sql string) (string, []columnDef, map[string]string, error) {
	matches := reCreateTable.FindStringSubmatch(sql)
	if matches == nil {
		return "", nil, nil, fmt.Errorf("SQL 语法错误：无法解析 CREATE TABLE 语句。\n期望格式：CREATE TABLE <表名> (<列名> <类型>, ...) [SETTINGS (<键>=<值>, ...)]")
	}

	tableName := strings.TrimSpace(matches[1])
	columnsPart := strings.TrimSpace(matches[2])
	settingsPart := strings.TrimSpace(matches[3])

	if tableName == "" {
		return "", nil, nil, fmt.Errorf("SQL 语法错误：缺少表名")
	}

	// 解析列定义
	columns, err := parseColumns(columnsPart)
	if err != nil {
		return "", nil, nil, err
	}

	// 解析 SETTINGS（可选）
	var settings map[string]string
	if settingsPart != "" {
		settings, err = parseSettings(settingsPart)
		if err != nil {
			return "", nil, nil, err
		}
	}

	return tableName, columns, settings, nil
}

// parseColumns 解析列定义部分，如 "name TEXT, age INTEGER"。
func parseColumns(columnsPart string) ([]columnDef, error) {
	if columnsPart == "" {
		return nil, fmt.Errorf("SQL 语法错误：列定义不能为空")
	}

	parts := strings.Split(columnsPart, ",")
	columns := make([]columnDef, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tokens := strings.Fields(part)
		if len(tokens) < 2 {
			return nil, fmt.Errorf("SQL 语法错误：列定义 '%s' 格式不正确，期望格式：<列名> <类型>", part)
		}

		colName := tokens[0]
		sqlType := strings.ToUpper(tokens[1])

		esType, ok := sqlTypeMapping[sqlType]
		if !ok {
			supported := make([]string, 0, len(sqlTypeMapping))
			for k := range sqlTypeMapping {
				supported = append(supported, k)
			}
			return nil, fmt.Errorf("SQL 语法错误：不支持的数据类型 '%s'。支持的类型：%s", tokens[1], strings.Join(supported, ", "))
		}

		columns = append(columns, columnDef{
			Name: colName,
			Type: esType,
		})
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("SQL 语法错误：至少需要定义一个列")
	}

	return columns, nil
}

// parseSettings 解析 SETTINGS 部分，如 "number_of_shards=3, number_of_replicas=1"。
func parseSettings(settingsPart string) (map[string]string, error) {
	settings := make(map[string]string)

	pairs := strings.Split(settingsPart, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		eqIdx := strings.Index(pair, "=")
		if eqIdx < 0 {
			return nil, fmt.Errorf("SQL 语法错误：SETTINGS 项 '%s' 格式不正确，期望格式：<键>=<值>", pair)
		}

		key := strings.TrimSpace(pair[:eqIdx])
		value := strings.TrimSpace(pair[eqIdx+1:])

		if key == "" {
			return nil, fmt.Errorf("SQL 语法错误：SETTINGS 键不能为空")
		}

		settings[key] = value
	}

	return settings, nil
}

// buildCreateIndexBody 构造 ES 创建索引的 JSON 请求体。
func buildCreateIndexBody(columns []columnDef, settings map[string]string) (string, error) {
	// 构造 mappings.properties
	properties := make(map[string]map[string]string, len(columns))
	for _, col := range columns {
		properties[col.Name] = map[string]string{
			"type": col.Type,
		}
	}

	body := map[string]any{
		"mappings": map[string]any{
			"properties": properties,
		},
	}

	// 添加 settings（如果有）
	if len(settings) > 0 {
		body["settings"] = settings
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("构造 JSON 请求体失败: %w", err)
	}

	return string(jsonBytes), nil
}
