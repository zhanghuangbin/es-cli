package executor

import (
	"context"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/formatter"
	"github.com/zhanghuangbin/es-cli/internal/translator"
)

type Executor struct {
	translator translator.Translator
	formatter  formatter.Formatter
	output     io.Writer
	handlers   map[SQLType]Handler
}

func New(t translator.Translator, f formatter.Formatter, output io.Writer, client *elasticsearch.Client) *Executor {
	handlers := map[SQLType]Handler{
		SQLTypeInsert: NewInsertHandler(client),
		SQLTypeUpdate: NewUpdateHandler(client),
		SQLTypeDelete: NewDeleteHandler(client),
		SQLTypeCreate: NewCreateHandler(client),
		SQLTypeDrop:   NewDropHandler(client),
		SQLTypeAlter:  NewAlterHandler(client),
	}

	return &Executor{
		translator: t,
		formatter:  f,
		output:     output,
		handlers:   handlers,
	}
}

func (e *Executor) SetFormatter(f formatter.Formatter) {
	e.formatter = f
}

func (e *Executor) Execute(sql string) error {
	sqlType := DetectSQLType(sql)

	var result *translator.Result
	var err error

	if sqlType == SQLTypeSelect {
		// SELECT 及其他未识别的语句走原有 _sql API 逻辑
		result, err = e.translator.Execute(context.Background(), sql)
	} else {
		handler, ok := e.handlers[sqlType]
		if !ok {
			return fmt.Errorf("不支持的 SQL 类型: %s", sqlType)
		}
		result, err = handler.Execute(context.Background(), sql)
	}

	if err != nil {
		return err
	}
	return e.formatter.Format(result, e.output)
}
