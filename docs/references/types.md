# types 模块

> `internal/types/`

全局共享类型。

- `sqltype.go` — `SQLType` 枚举（SELECT/INSERT/UPDATE/DELETE/CREATE/DROP/ALTER）+ `DetectSQLType()` 检测函数（未知类型 fallback 到 SELECT）
- `result.go` — `Result`（Columns + Rows + Source）和 `Meta`（Status + Type + Message + Endpoint + Stat）
