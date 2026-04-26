package cmd

import (
	"fmt"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

type Options struct {
	Format   string
	JSONPath string
	Template string
	Fields   []string
}

func Format(result *types.Result, w io.Writer, opts Options) error {
	switch opts.Format {
	case "json":
		return formatJSON(result, w, opts.JSONPath)
	case "csv":
		return formatCSV(result, w, opts.Fields)
	case "yaml":
		return formatYAML(result, w)
	case "go-template":
		return formatGoTemplate(result, w, opts.Template)
	default:
		return fmt.Errorf("不支持的输出格式: %s（可选: json, csv, yaml, go-template）", opts.Format)
	}
}
