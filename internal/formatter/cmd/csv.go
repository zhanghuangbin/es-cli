package cmd

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

type csvFormatter struct {
	fields []string
}

func (f *csvFormatter) Format(result *types.Result, w io.Writer) error {
	if len(result.Columns) == 0 {
		return nil
	}

	cols, indices := filterColumns(result.Columns, f.fields)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write(cols); err != nil {
		return fmt.Errorf("写入 CSV 表头失败: %w", err)
	}

	for _, row := range result.Rows {
		record := make([]string, len(indices))
		for i, idx := range indices {
			if idx < len(row) {
				record[i] = fmt.Sprintf("%v", row[idx])
			}
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入 CSV 行失败: %w", err)
		}
	}

	return nil
}

func filterColumns(columns []string, fields []string) ([]string, []int) {
	if len(fields) == 0 {
		indices := make([]int, len(columns))
		for i := range columns {
			indices[i] = i
		}
		return columns, indices
	}

	colIndex := make(map[string]int, len(columns))
	for i, col := range columns {
		colIndex[col] = i
	}

	var filtered []string
	var indices []int
	for _, f := range fields {
		if idx, ok := colIndex[f]; ok {
			filtered = append(filtered, f)
			indices = append(indices, idx)
		}
	}

	return filtered, indices
}
