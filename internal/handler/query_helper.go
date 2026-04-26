package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/types"
	"github.com/zhanghuangbin/es-cli/pkg/es"
	"sort"
)

// fetchDocByID 根据文档 ID 回查单条文档，返回包含 _id 和 _source 所有字段的 Result。
func fetchDocByID(ctx context.Context, client *elasticsearch.Client, index, docID string) (*types.Result, error) {
	path := fmt.Sprintf("/%s/_doc/%s", index, docID)
	respBody, err := es.DoRequest(ctx, client, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("回查文档失败: %w", err)
	}

	var doc struct {
		ID     string         `json:"_id"`
		Source map[string]any `json:"_source"`
	}
	if err := json.Unmarshal(respBody, &doc); err != nil {
		return nil, fmt.Errorf("解析文档响应失败: %w", err)
	}

	return buildResultFromSource(doc.ID, doc.Source), nil
}

// buildResultFromSource 将单条文档的 _id 和 _source 构造为 Result。
func buildResultFromSource(id string, source map[string]any) *types.Result {
	// 字段名排序，保证输出稳定
	columns := make([]string, 0, len(source))
	for k := range source {
		columns = append(columns, k)
	}
	sort.Strings(columns)
	// _id 放在最前面
	columns = append([]string{"_id"}, columns...)

	row := make([]any, len(columns))
	row[0] = id
	for i, col := range columns[1:] {
		row[i+1] = source[col]
	}

	return &types.Result{
		Meta: types.Meta{
			Status:  200,
			Message: "操作成功",
		},
		Columns: columns,
		Rows:    [][]any{row},
	}
}
