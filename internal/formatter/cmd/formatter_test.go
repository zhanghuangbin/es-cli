package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

func TestNewFormatterFactory(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{"json", Options{Format: "json"}, false},
		{"csv", Options{Format: "csv"}, false},
		{"yaml", Options{Format: "yaml"}, false},
		{"go-template with template", Options{Format: "go-template", Template: "{{.Columns}}"}, false},
		{"go-template without template", Options{Format: "go-template"}, true},
		{"unknown format", Options{Format: "xml"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFormatter(tt.opts)
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

func TestJsonFormatterWithSource(t *testing.T) {
	result := &types.Result{
		Source: `{"name":"test","age":25}`,
	}

	f, _ := NewFormatter(Options{Format: "json"})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name") {
		t.Error("output should contain 'name'")
	}
}

func TestJsonFormatterWithoutSource(t *testing.T) {
	result := &types.Result{
		Columns: []string{"name", "age"},
		Rows:    [][]any{{"test", 25}},
	}

	f, _ := NewFormatter(Options{Format: "json"})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证输出是有效的 JSON 数组
	var data []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(data) != 1 {
		t.Errorf("expected 1 row, got %d", len(data))
	}
}

func TestJsonFormatterWithJSONPath(t *testing.T) {
	result := &types.Result{
		Source: `{"data":{"items":[1,2,3]}}`,
	}

	f, _ := NewFormatter(Options{Format: "json", JSONPath: "$.data.items"})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1") {
		t.Error("output should contain extracted data")
	}
}

func TestCsvFormatter(t *testing.T) {
	result := &types.Result{
		Columns: []string{"name", "age", "city"},
		Rows: [][]any{
			{"张三", 25, "北京"},
			{"李四", 30, "上海"},
		},
	}

	f, _ := NewFormatter(Options{Format: "csv"})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 { // header + 2 rows
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "name") {
		t.Error("header should contain 'name'")
	}
}

func TestCsvFormatterWithFields(t *testing.T) {
	result := &types.Result{
		Columns: []string{"name", "age", "city"},
		Rows: [][]any{
			{"张三", 25, "北京"},
		},
	}

	f, _ := NewFormatter(Options{Format: "csv", Fields: []string{"name", "city"}})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// 表头应只有 name 和 city
	if strings.Contains(lines[0], "age") {
		t.Error("header should not contain 'age' when filtered")
	}
}

func TestCsvFormatterEmptyColumns(t *testing.T) {
	result := &types.Result{
		Columns: []string{},
		Rows:    [][]any{},
	}

	f, _ := NewFormatter(Options{Format: "csv"})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestYamlFormatter(t *testing.T) {
	result := &types.Result{
		Columns: []string{"name", "age"},
		Rows:    [][]any{{"test", 25}},
	}

	f, _ := NewFormatter(Options{Format: "yaml"})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name") {
		t.Error("output should contain 'name'")
	}
	if !strings.Contains(output, "test") {
		t.Error("output should contain 'test'")
	}
}

func TestGoTemplateFormatter(t *testing.T) {
	result := &types.Result{
		Columns: []string{"name"},
		Rows:    [][]any{{"test"}},
	}

	f, _ := NewFormatter(Options{Format: "go-template", Template: "列数: {{len .Columns}}"})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "列数: 1" {
		t.Errorf("output = %q, want '列数: 1'", output)
	}
}

func TestGoTemplateFormatterInvalidTemplate(t *testing.T) {
	result := &types.Result{}

	f, _ := NewFormatter(Options{Format: "go-template", Template: "{{.Invalid"})
	var buf bytes.Buffer
	err := f.Format(result, &buf)
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestFilterColumns(t *testing.T) {
	tests := []struct {
		name        string
		columns     []string
		fields      []string
		wantCols    []string
		wantIndices []int
	}{
		{
			name:        "no filter",
			columns:     []string{"a", "b", "c"},
			fields:      nil,
			wantCols:    []string{"a", "b", "c"},
			wantIndices: []int{0, 1, 2},
		},
		{
			name:        "filter some",
			columns:     []string{"a", "b", "c"},
			fields:      []string{"a", "c"},
			wantCols:    []string{"a", "c"},
			wantIndices: []int{0, 2},
		},
		{
			name:        "filter non-existent",
			columns:     []string{"a", "b"},
			fields:      []string{"x", "y"},
			wantCols:    nil,
			wantIndices: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, indices := filterColumns(tt.columns, tt.fields)
			if len(cols) != len(tt.wantCols) {
				t.Fatalf("cols count = %d, want %d", len(cols), len(tt.wantCols))
			}
			for i := range tt.wantCols {
				if cols[i] != tt.wantCols[i] {
					t.Errorf("cols[%d] = %q, want %q", i, cols[i], tt.wantCols[i])
				}
			}
			if len(indices) != len(tt.wantIndices) {
				t.Fatalf("indices count = %d, want %d", len(indices), len(tt.wantIndices))
			}
			for i := range tt.wantIndices {
				if indices[i] != tt.wantIndices[i] {
					t.Errorf("indices[%d] = %d, want %d", i, indices[i], tt.wantIndices[i])
				}
			}
		})
	}
}

func TestRowsToMaps(t *testing.T) {
	result := &types.Result{
		Columns: []string{"name", "age"},
		Rows: [][]any{
			{"张三", 25},
			{"李四", 30},
		},
	}

	maps := rowsToMaps(result)
	if len(maps) != 2 {
		t.Fatalf("expected 2 maps, got %d", len(maps))
	}
	if maps[0]["name"] != "张三" {
		t.Errorf("maps[0]['name'] = %v, want '张三'", maps[0]["name"])
	}
	if maps[1]["age"] != 30 {
		t.Errorf("maps[1]['age'] = %v, want 30", maps[1]["age"])
	}
}

func TestRowsToMapsRowShorterThanColumns(t *testing.T) {
	result := &types.Result{
		Columns: []string{"name", "age", "city"},
		Rows: [][]any{
			{"张三", 25}, // 缺少 city
		},
	}

	maps := rowsToMaps(result)
	if len(maps) != 1 {
		t.Fatalf("expected 1 map, got %d", len(maps))
	}
	if _, ok := maps[0]["city"]; ok {
		t.Error("city should not be in map when row is shorter than columns")
	}
}
