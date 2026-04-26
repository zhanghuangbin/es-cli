package formatter

import (
	"io"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

// Formatter 是所有输出格式化器的通用接口。
type Formatter interface {
	Format(result *types.Result, w io.Writer) error
}
