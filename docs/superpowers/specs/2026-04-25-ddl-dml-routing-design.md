# DDL/DML SQL 混合路由

## 背景

es-cli 的 MVP 版本将所有 SQL 语句直接转发给 Elasticsearch 的 `/_sql` API。但 `/_sql` API 不支持 DDL（如 `CREATE INDEX`）和大部分 DML 操作，限制了工具的实用性。

本方案通过轻量 SQL 类型检测，将不同 SQL 路由到不同处理器：SELECT 继续走 `_sql` API，DDL/DML 直接调用 ES REST API。

## 架构

```
REPL.executeInput()
  → executor.Execute(sql)
    → SQLTypeDetector.detect(sql)
    → RoutingExecutor.dispatch(sql, type)
        ├── selectHandler    → BuiltinTranslator → POST /_sql
        ├── createHandler    → PUT /{index} + mappings + settings
        ├── dropHandler      → DELETE /{index}
        ├── insertHandler    → POST /{index}/_doc
        ├── updateHandler    → POST /{index}/_update/{id}
        ├── deleteHandler    → POST /{index}/_delete_by_query
        └── alterHandler     → POST /{index}/_settings 或 _rename 别名
```

## SQL 语法

### DDL

```sql
CREATE TABLE users (name TEXT, age INTEGER) SETTINGS (number_of_shards = 3);
DROP TABLE users;
ALTER INDEX users SETTINGS (number_of_replicas = 2);
ALTER TABLE old_index RENAME TO new_index;
```

### DML

```sql
INSERT INTO users (name, age) VALUES ('alice', 25);
UPDATE users SET age = 26 WHERE name = 'alice';
DELETE FROM users WHERE name = 'alice';
```

### 数据类型映射

| SQL 类型 | ES 类型 |
|----------|---------|
| TEXT | text |
| INTEGER | integer |
| LONG | long |
| FLOAT | float |
| DOUBLE | double |
| BOOLEAN | boolean |
| DATE | date |
| KEYWORD | keyword |

## 文件变更

| 文件 | 动作 |
|------|------|
| `internal/executor/executor.go` | 修改：增加路由逻辑 |
| `internal/executor/handler.go` | 新增：Handler 接口 |
| `internal/executor/sqltype.go` | 新增：SQL 类型检测器 |
| `internal/executor/create_handler.go` | 新增 |
| `internal/executor/drop_handler.go` | 新增 |
| `internal/executor/insert_handler.go` | 新增 |
| `internal/executor/update_handler.go` | 新增 |
| `internal/executor/delete_handler.go` | 新增 |
| `internal/executor/alter_handler.go` | 新增 |
| `internal/repl/completer.go` | 修改：补充 DDL/DML 关键词 |

## 解析策略

轻量级：正则 + 字符串操作，不引入完整 SQL 解析器。每个 handler 自行解析其语句的结构化部分。

## 错误处理

- 解析失败：明确提示语法错误位置
- ES API 错误：透传 ES 返回的错误信息
- 未知 SQL 类型：fallback 到 `/_sql` API
