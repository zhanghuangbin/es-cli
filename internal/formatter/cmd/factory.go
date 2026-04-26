package cmd

import (
	"fmt"

	"github.com/zhanghuangbin/es-cli/internal/formatter"
)

// Options 定义 cmd 格式化器的配置选项。
type Options struct {
	Format   string
	JSONPath string
	Template string
	Fields   []string
}

// NewFormatter 根据选项创建对应的 cmd 格式化器。
func NewFormatter(opts Options) (formatter.Formatter, error) {
	switch opts.Format {
	case "json":
		return &jsonFormatter{jsonPath: opts.JSONPath}, nil
	case "csv":
		return &csvFormatter{fields: opts.Fields}, nil
	case "yaml":
		return &yamlFormatter{}, nil
	case "go-template":
		if opts.Template == "" {
			return nil, fmt.Errorf("使用 go-template 格式时必须指定 --template 参数")
		}
		return &goTemplateFormatter{template: opts.Template}, nil
	default:
		return nil, fmt.Errorf("不支持的输出格式: %s（可选: json, csv, yaml, go-template）", opts.Format)
	}
}
