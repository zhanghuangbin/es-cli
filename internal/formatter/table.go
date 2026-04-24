package formatter

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/zhanghuangbin/es-cli/internal/translator"
)

type TableFormatter struct{}

func (f *TableFormatter) Format(result *translator.Result, w io.Writer) error {
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
	return nil
}
