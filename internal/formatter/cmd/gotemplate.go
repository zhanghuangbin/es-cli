package cmd

import (
	"fmt"
	"io"
	"text/template"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

type goTemplateFormatter struct {
	template string
}

func (f *goTemplateFormatter) Format(result *types.Result, w io.Writer) error {
	t, err := template.New("output").Parse(f.template)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	return t.Execute(w, result)
}
