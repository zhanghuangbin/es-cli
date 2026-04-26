package formatter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/zhanghuangbin/es-cli/internal/types"
)

type TableFormatter struct{}

func (f *TableFormatter) Format(result *types.Result, w io.Writer) error {
	if len(result.Columns) == 0 {
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(w)

	header := make(table.Row, len(result.Columns))
	for i, col := range result.Columns {
		header[i] = col
	}
	t.AppendHeader(header)

	for _, row := range result.Rows {
		tableRow := make(table.Row, len(row))
		for i, val := range row {
			tableRow[i] = val
		}
		t.AppendRow(tableRow)
	}

	t.SetStyle(table.StyleLight)
	t.Render()

	// 打印统计信息
	f.renderStat(result.Meta.Stat, w)

	return nil
}

// renderStat 在表格下方打印统计信息。
func (f *TableFormatter) renderStat(stat map[string]any, w io.Writer) {
	if len(stat) == 0 {
		return
	}

	// 按 key 排序，保证输出稳定
	keys := make([]string, 0, len(stat))
	for k := range stat {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %v", k, stat[k]))
	}

	fmt.Fprintf(w, "\n%s\n", strings.Join(parts, ", "))
}
