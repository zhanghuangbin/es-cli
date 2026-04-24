package executor

import (
	"context"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/formatter"
	"github.com/zhanghuangbin/es-cli/internal/translator"
)

type Executor struct {
	translator translator.Translator
	formatter  formatter.Formatter
	output     io.Writer
}

func New(t translator.Translator, f formatter.Formatter, output io.Writer) *Executor {
	return &Executor{
		translator: t,
		formatter:  f,
		output:     output,
	}
}

func (e *Executor) SetFormatter(f formatter.Formatter) {
	e.formatter = f
}

func (e *Executor) Execute(sql string) error {
	result, err := e.translator.Execute(context.Background(), sql)
	if err != nil {
		return err
	}
	return e.formatter.Format(result, e.output)
}
