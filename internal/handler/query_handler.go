package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/types"
)

// QueryHandler 处理 SELECT 及其他未识别的 SQL 语句，通过 ES _sql API 执行。
type QueryHandler struct {
	client *elasticsearch.Client
}

// NewQueryHandler 创建一个新的 QueryHandler 实例。
func NewQueryHandler(client *elasticsearch.Client) *QueryHandler {
	return &QueryHandler{client: client}
}

type sqlRequest struct {
	Query string `json:"query"`
}

type sqlResponse struct {
	Columns []sqlColumn     `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

type sqlColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (h *QueryHandler) Execute(ctx context.Context, sql string) (*types.Result, error) {
	body, err := json.Marshal(sqlRequest{Query: sql})
	if err != nil {
		return nil, fmt.Errorf("序列化 SQL 请求失败: %w", err)
	}

	res, err := h.client.SQL.Query(
		bytes.NewReader(body),
		h.client.SQL.Query.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("执行 SQL 失败: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if res.IsError() {
		return &types.Result{
			Meta: types.Meta{Status: res.StatusCode, Message: string(respBody)},
		}, fmt.Errorf("ES 错误 [%d]: %s", res.StatusCode, respBody)
	}

	var sqlRes sqlResponse
	if err := json.Unmarshal(respBody, &sqlRes); err != nil {
		return nil, fmt.Errorf("解析 SQL 响应失败: %w", err)
	}

	columns := make([]string, len(sqlRes.Columns))
	for i, col := range sqlRes.Columns {
		columns[i] = col.Name
	}

	rows := make([][]any, len(sqlRes.Rows))
	for i, row := range sqlRes.Rows {
		rows[i] = row
	}

	return &types.Result{
		Meta: types.Meta{
			Status: res.StatusCode,
			Stat: map[string]any{
				"返回行数": len(rows),
			},
		},
		Columns: columns,
		Rows:    rows,
	}, nil
}
