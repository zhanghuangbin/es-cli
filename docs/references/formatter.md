# formatter 模块

> `internal/formatter/`

输出格式化。`Formatter` 接口定义在 `formatter.go`，通过 `New(format)` 工厂函数创建。

| 格式 | 实现 | 状态 |
|------|------|------|
| `table` | `TableFormatter`（`table.go`，基于 `go-pretty`） | 默认，含统计信息输出 |
| `json` | `JsonFormatter`（`json.go`） | 输出 ES 接口路径 + 原始响应体（缩进格式化） |
| `csv` | — | 未实现 |
