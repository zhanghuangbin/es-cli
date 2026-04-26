package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

// JsonFormatter 以 JSON 格式输出结果，包含 ES REST 接口和原始响应体。
type JsonFormatter struct{}

func (f *JsonFormatter) Format(result *types.Result, w io.Writer) error {
	// 打印 ES REST 接口
	if result.Meta.Endpoint != "" {
		fmt.Fprintf(w, "接口: %s\n", result.Meta.Endpoint)
	}

	// 打印 ES 原始响应体（格式化缩进）
	if result.Source != "" {
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(result.Source), "", "  "); err != nil {
			// 格式化失败则直接输出原始数据
			fmt.Fprint(w, result.Source)
		} else {
			buf.WriteTo(w)
		}
		fmt.Fprintln(w)
	}

	return nil
}
