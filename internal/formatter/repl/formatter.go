package repl

import (
	"fmt"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

type Formatter interface {
	Format(result *types.Result, w io.Writer) error
}

func New(format string) (Formatter, error) {
	switch format {
	case "table":
		return &TableFormatter{}, nil
	case "json":
		return &JsonFormatter{}, nil
	default:
		return nil, fmt.Errorf("未知格式: %s（可选: table, json）", format)
	}
}
