# 实现计划：DDL/DML SQL 混合路由

## 步骤 1：核心基础设施

### 1.1 新增 `internal/executor/sqltype.go`
- 定义 `SQLType` 枚举（Select/Insert/Update/Delete/Create/Drop/Alter）
- 实现 `Detect(sql string) SQLType`：用 `strings.HasPrefix` + `strings.TrimSpace` 判断语句类型
- 未知类型 fallback 到 Select（走 `_sql` API）

### 1.2 新增 `internal/executor/handler.go`
- 定义 `Handler` 接口：`Execute(ctx context.Context, sql string) (*translator.Result, error)`
- 每个 handler 实现此接口

### 1.3 新增 `pkg/es/rest.go`
- 封装通用 HTTP 辅助函数：`doRequest(client, method, path, body) ([]byte, error)`
- 用于构造 ES REST API 请求和解析响应

## 步骤 2：实现各 Handler

### 2.1 `internal/executor/create_handler.go`
- 解析 `CREATE TABLE name (col TYPE, ...) SETTINGS (k=v, ...)`
- SQL 类型 → ES 类型映射
- 调用 `PUT /{index}` with mappings + settings
- 返回成功消息

### 2.2 `internal/executor/drop_handler.go`
- 解析 `DROP TABLE name`
- 调用 `DELETE /{name}`

### 2.3 `internal/executor/insert_handler.go`
- 解析 `INSERT INTO name (cols) VALUES (vals)`
- 调用 `POST /{name}/_doc`

### 2.4 `internal/executor/update_handler.go`
- 解析 `UPDATE name SET k=v WHERE condition`
- 调用 `POST /{name}/_update_by_query`

### 2.5 `internal/executor/delete_handler.go`
- 解析 `DELETE FROM name WHERE condition`
- 调用 `POST /{name}/_delete_by_query`

### 2.6 `internal/executor/alter_handler.go`
- 解析 `ALTER INDEX name SETTINGS (...))` → `POST /{name}/_settings`
- 解析 `ALTER TABLE old RENAME TO new` → `_aliases` 交换

## 步骤 3：集成到 Executor

### 3.1 修改 `internal/executor/executor.go`
- 添加 ES client 字段
- `Execute()` 中调用 `Detect()` 获取 SQL 类型
- 根据类型路由到对应 Handler
- Select 类型保持原有 `BuiltinTranslator` 逻辑

## 步骤 4：完善 REPL 体验

### 4.1 修改 `internal/repl/completer.go`
- 补充 DDL 关键词：CREATE TABLE, DROP TABLE, ALTER INDEX, ALTER TABLE, RENAME TO, SETTINGS

### 4.2 修改 `internal/repl/repl.go`
- `.help` 中增加 DDL/DML 示例

## 步骤 5：测试与验收

### 5.1 单元测试
- `sqltype_test.go`：验证各类型 SQL 的检测
- `create_handler_test.go`：验证 CREATE TABLE 解析

### 5.2 编译验收
- `go build ./...` 无错误
- `go test ./...` 通过

## 实现顺序

1. sqltype.go → handler.go → rest.go（基础设施）
2. 逐个实现 handler（create → drop → insert → update → delete → alter）
3. 修改 executor.go 集成路由
4. 修改 completer.go 和 repl.go
5. 测试与验收
