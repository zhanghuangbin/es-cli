# executor 模块

> `internal/executor/executor.go`

SQL 调度器。通过 `types.DetectSQLType()` 识别 SQL 类型，分发到对应 Handler，将结果交给 Formatter 输出。

- `New(f, output, client)` — 创建实例，内部注册所有 Handler
- `SetFormatter(f)` — 动态切换 Formatter
- `Execute(sql)` — 解析 → 分发 → 格式化输出
