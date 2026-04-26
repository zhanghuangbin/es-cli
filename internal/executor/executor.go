package executor

import (
	"context"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/formatter"
	"github.com/zhanghuangbin/es-cli/internal/handler"
	"github.com/zhanghuangbin/es-cli/internal/types"
)

// Executor 是 SQL 执行调度器，根据 SQL 类型分发到对应的 Handler。
type Executor struct {
	formatter formatter.Formatter
	output    io.Writer
	handlers  map[types.SQLType]handler.Handler
}

// New 创建一个新的 Executor 实例，内部注册所有 Handler。
func New(f formatter.Formatter, output io.Writer, client *elasticsearch.Client) *Executor {
	handlers := map[types.SQLType]handler.Handler{
		types.SQLTypeSelect: handler.NewQueryHandler(client),
		types.SQLTypeInsert: handler.NewInsertHandler(client),
		types.SQLTypeUpdate: handler.NewUpdateHandler(client),
		types.SQLTypeDelete: handler.NewDeleteHandler(client),
		types.SQLTypeCreate: handler.NewCreateHandler(client),
		types.SQLTypeDrop:   handler.NewDropHandler(client),
		types.SQLTypeAlter:  handler.NewAlterHandler(client),
	}

	return &Executor{
		formatter: f,
		output:    output,
		handlers:  handlers,
	}
}

// SetFormatter 动态切换输出格式化器。
func (e *Executor) SetFormatter(f formatter.Formatter) {
	e.formatter = f
}

// Execute 解析 SQL 类型并分发到对应的 Handler 执行。
func (e *Executor) Execute(sql string) error {
	sqlType := types.DetectSQLType(sql)

	h, ok := e.handlers[sqlType]
	if !ok {
		return fmt.Errorf("不支持的 SQL 类型: %s", sqlType)
	}

	result, err := h.Execute(context.Background(), sql)
	if err != nil {
		return err
	}

	// 设置 SQL 类型，方便 formatter 根据类型决定渲染方式
	result.Meta.Type = sqlType

	return e.formatter.Format(result, e.output)
}
