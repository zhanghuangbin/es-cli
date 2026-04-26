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

// AlterHandler 处理 ALTER 语句，支持修改索引设置和重命名索引（通过别名）。
type AlterHandler struct {
	client *elasticsearch.Client
}

// NewAlterHandler 创建一个新的 AlterHandler 实例。
func NewAlterHandler(client *elasticsearch.Client) *AlterHandler {
	return &AlterHandler{client: client}
}

// reAlterSettings 匹配 ALTER INDEX <name> SETTINGS (...) 语句。
// 捕获组：1=索引名, 2=settings 内容
var reAlterSettings = regexp.MustCompile(
	`(?i)^\s*ALTER\s+INDEX\s+(\S+)\s+SETTINGS\s*\(\s*(.*?)\s*\)\s*;?\s*$`,
)

// reAlterRename 匹配 ALTER TABLE <old> RENAME TO <new> 语句。
// 捕获组：1=旧表名, 2=新表名
var reAlterRename = regexp.MustCompile(
	`(?i)^\s*ALTER\s+TABLE\s+(\S+)\s+RENAME\s+TO\s+(\S+)\s*;?\s*$`,
)

// Execute 解析 ALTER 语句并执行对应的 ES API 操作。
//
// 支持的语法：
//
//	ALTER INDEX <索引名> SETTINGS (<键1>=<值1>, <键2>=<值2>, ...)
//	ALTER TABLE <旧表名> RENAME TO <新表名>
func (h *AlterHandler) Execute(ctx context.Context, sql string) (*types.Result, error) {
	// 尝试匹配 ALTER INDEX ... SETTINGS (...)
	if matches := reAlterSettings.FindStringSubmatch(sql); matches != nil {
		return h.executeAlterSettings(ctx, matches[1], matches[2])
	}

	// 尝试匹配 ALTER TABLE ... RENAME TO ...
	if matches := reAlterRename.FindStringSubmatch(sql); matches != nil {
		return h.executeAlterRename(ctx, matches[1], matches[2])
	}

	return nil, fmt.Errorf("SQL 语法错误：无法解析 ALTER 语句。\n支持的格式：\n  ALTER INDEX <索引名> SETTINGS (<键>=<值>, ...)\n  ALTER TABLE <旧表名> RENAME TO <新表名>")
}

// executeAlterSettings 执行 ALTER INDEX ... SETTINGS 操作。
// 调用 ES PUT /{index}/_settings API 更新索引设置。
func (h *AlterHandler) executeAlterSettings(ctx context.Context, indexName, settingsPart string) (*types.Result, error) {
	indexName = strings.TrimSpace(indexName)
	settingsPart = strings.TrimSpace(settingsPart)

	if indexName == "" {
		return nil, fmt.Errorf("SQL 语法错误：缺少索引名")
	}

	// 复用 create_handler.go 中的 parseSettings 函数
	settings, err := parseSettings(settingsPart)
	if err != nil {
		return nil, err
	}

	if len(settings) == 0 {
		return nil, fmt.Errorf("SQL 语法错误：SETTINGS 不能为空")
	}

	// 构造请求体
	body := map[string]any{
		"settings": settings,
	}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("构造 JSON 请求体失败: %w", err)
	}

	// 调用 ES API 更新设置
	path := fmt.Sprintf("/%s/_settings", indexName)
	_, err = es.DoRequest(ctx, h.client, "PUT", path, strings.NewReader(string(jsonBytes)))
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("索引 %s 设置更新成功", indexName)
	return &types.Result{
		Meta: types.Meta{
			Status:  200,
			Message: msg,
		},
		Columns: []string{"结果"},
		Rows:    [][]any{{msg}},
	}, nil
}

// executeAlterRename 执行 ALTER TABLE ... RENAME TO 操作。
// 通过 ES POST /_aliases API 为旧索引添加别名，实现逻辑重命名。
// 注意：这不是真正的重命名，而是添加别名。
func (h *AlterHandler) executeAlterRename(ctx context.Context, oldName, newName string) (*types.Result, error) {
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)

	if oldName == "" {
		return nil, fmt.Errorf("SQL 语法错误：缺少原表名")
	}
	if newName == "" {
		return nil, fmt.Errorf("SQL 语法错误：缺少新表名")
	}

	// 构造 _aliases 请求体：为旧索引添加别名
	body := map[string]any{
		"actions": []map[string]any{
			{
				"add": map[string]string{
					"index": oldName,
					"alias": newName,
				},
			},
		},
	}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("构造 JSON 请求体失败: %w", err)
	}

	// 调用 ES API 添加别名
	_, err = es.DoRequest(ctx, h.client, "POST", "/_aliases", strings.NewReader(string(jsonBytes)))
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("已为索引 %s 添加别名 %s", oldName, newName)
	return &types.Result{
		Meta: types.Meta{
			Status:  200,
			Message: msg,
		},
		Columns: []string{"结果"},
		Rows:    [][]any{{msg}},
	}, nil
}
