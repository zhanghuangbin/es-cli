package formatter

import (
	"fmt"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/translator"
)

type Formatter interface {
	Format(result *translator.Result, w io.Writer) error
}

func New(format string) (Formatter, error) {
	switch format {
	case "table":
		return &TableFormatter{}, nil
	case "json", "csv":
		return nil, fmt.Errorf("格式 '%s' 暂未实现，敬请期待", format)
	default:
		return nil, fmt.Errorf("未知格式: %s（可选: table, json, csv）", format)
	}
}
