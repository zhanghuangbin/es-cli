package cmd

import (
	"fmt"
	"io"
	"text/template"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

func formatGoTemplate(result *types.Result, w io.Writer, tmpl string) error {
	if tmpl == "" {
		return fmt.Errorf("使用 go-template 格式时必须指定 --template 参数")
	}

	t, err := template.New("output").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	return t.Execute(w, result)
}
