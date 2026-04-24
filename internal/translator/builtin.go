package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
)

type BuiltinTranslator struct {
	client *elasticsearch.Client
}

func NewBuiltinTranslator(client *elasticsearch.Client) *BuiltinTranslator {
	return &BuiltinTranslator{client: client}
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

func (t *BuiltinTranslator) Execute(ctx context.Context, sql string) (*Result, error) {
	body, err := json.Marshal(sqlRequest{Query: sql})
	if err != nil {
		return nil, fmt.Errorf("序列化 SQL 请求失败: %w", err)
	}

	res, err := t.client.SQL.Query(
		bytes.NewReader(body),
		t.client.SQL.Query.WithContext(ctx),
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
		return &Result{
			Meta: Meta{Status: res.StatusCode, Message: string(respBody)},
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

	return &Result{
		Meta:    Meta{Status: res.StatusCode},
		Columns: columns,
		Rows:    rows,
	}, nil
}
