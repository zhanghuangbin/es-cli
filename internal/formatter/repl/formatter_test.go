package repl

import (
	"bytes"
	"strings"
	"testing"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"table", "table", false},
		{"json", "json", false},
		{"unknown", "xml", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFormatter(tt.format)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f == nil {
				t.Error("expected non-nil formatter")
			}
		})
	}
}

func TestTableFormatterFormat(t *testing.T) {
	result := &types.Result{
		Columns: []string{"name", "age"},
		Rows: [][]any{
			{"张三", 25},
			{"李四", 30},
		},
		Meta: types.Meta{
			Stat: map[string]any{
				"返回行数": 2,
			},
		},
	}

	var buf bytes.Buffer
	f := &TableFormatter{}
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// go-pretty 渲染的表格中包含列名（大写形式或原始形式）和数据
	if !strings.Contains(output, "NAME") && !strings.Contains(output, "name") {
		t.Error("output should contain column 'name'")
	}
	if !strings.Contains(output, "AGE") && !strings.Contains(output, "age") {
		t.Error("output should contain column 'age'")
	}
	if !strings.Contains(output, "张三") {
		t.Error("output should contain '张三'")
	}
	if !strings.Contains(output, "返回行数") {
		t.Error("output should contain stat '返回行数'")
	}
}

func TestTableFormatterEmptyColumns(t *testing.T) {
	result := &types.Result{
		Columns: []string{},
		Rows:    [][]any{},
	}

	var buf bytes.Buffer
	f := &TableFormatter{}
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestJsonFormatterFormat(t *testing.T) {
	result := &types.Result{
		Meta: types.Meta{
			Endpoint: "POST /_sql",
		},
		Source: `{"columns":[{"name":"id","type":"integer"}],"rows":[[1]]}`,
	}

	var buf bytes.Buffer
	f := &JsonFormatter{}
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "接口: POST /_sql") {
		t.Error("output should contain endpoint info")
	}
	if !strings.Contains(output, "columns") {
		t.Error("output should contain formatted JSON")
	}
}

func TestJsonFormatterNoSource(t *testing.T) {
	result := &types.Result{
		Meta: types.Meta{},
	}

	var buf bytes.Buffer
	f := &JsonFormatter{}
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 无 endpoint 和 source 时不应输出任何内容
}

func TestJsonFormatterInvalidJSON(t *testing.T) {
	result := &types.Result{
		Meta: types.Meta{
			Endpoint: "POST /_sql",
		},
		Source: `{invalid json}`,
	}

	var buf bytes.Buffer
	f := &JsonFormatter{}
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// 无效 JSON 应该原样输出
	if !strings.Contains(output, "{invalid json}") {
		t.Error("invalid JSON should be output as-is")
	}
}
