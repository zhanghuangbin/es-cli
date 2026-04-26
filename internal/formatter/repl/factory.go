package repl

import (
	"fmt"

	"github.com/zhanghuangbin/es-cli/internal/formatter"
)

// NewFormatter 根据格式名称创建对应的 REPL 格式化器。
func NewFormatter(format string) (formatter.Formatter, error) {
	switch format {
	case "table":
		return &TableFormatter{}, nil
	case "json":
		return &JsonFormatter{}, nil
	default:
		return nil, fmt.Errorf("未知格式: %s（可选: table, json）", format)
	}
}
