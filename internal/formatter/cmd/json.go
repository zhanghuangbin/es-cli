package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/PaesslerAG/jsonpath"
	"github.com/zhanghuangbin/es-cli/internal/types"
)

type jsonFormatter struct {
	jsonPath string
}

func (f *jsonFormatter) Format(result *types.Result, w io.Writer) error {
	if f.jsonPath != "" {
		return f.formatJSONPath(result, w)
	}

	if result.Source != "" {
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(result.Source), "", "  "); err != nil {
			fmt.Fprint(w, result.Source)
		} else {
			buf.WriteTo(w)
		}
		fmt.Fprintln(w)
		return nil
	}

	data := rowsToMaps(result)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *jsonFormatter) formatJSONPath(result *types.Result, w io.Writer) error {
	var raw any
	if result.Source != "" {
		if err := json.Unmarshal([]byte(result.Source), &raw); err != nil {
			return fmt.Errorf("解析 JSON 失败: %w", err)
		}
	} else {
		raw = map[string]any{
			"columns": result.Columns,
			"rows":    result.Rows,
		}
	}

	extracted, err := jsonpath.Get(f.jsonPath, raw)
	if err != nil {
		return fmt.Errorf("JSONPath 提取失败: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(extracted)
}

func rowsToMaps(result *types.Result) []map[string]any {
	maps := make([]map[string]any, 0, len(result.Rows))
	for _, row := range result.Rows {
		m := make(map[string]any, len(result.Columns))
		for i, col := range result.Columns {
			if i < len(row) {
				m[col] = row[i]
			}
		}
		maps = append(maps, m)
	}
	return maps
}
