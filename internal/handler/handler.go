package handler

import (
	"context"

	"github.com/zhanghuangbin/es-cli/internal/types"
)

// Handler 定义了执行 SQL 语句的接口。
// 每种 SQLType 对应一个具体的 Handler 实现。
type Handler interface {
	// Execute 执行 SQL 语句并返回查询结果。
	Execute(ctx context.Context, sql string) (*types.Result, error)
}
