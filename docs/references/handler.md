# handler 模块

> `internal/handler/`

每种 SQL 类型一个 Handler，均实现 `Handler` 接口（`handler.go`）。

| Handler | SQL | ES API |
|---------|-----|--------|
| QueryHandler | SELECT / SHOW TABLES / DESCRIBE | `POST /_sql` |
| CreateHandler | CREATE TABLE ... [SETTINGS ...] | `PUT /<idx>` |
| InsertHandler | INSERT INTO ... VALUES ... | `POST /<idx>/_doc` |
| UpdateHandler | UPDATE ... SET ... WHERE ... | `POST /<idx>/_update_by_query` |
| DeleteHandler | DELETE FROM ... WHERE ... | `POST /<idx>/_delete_by_query` |
| DropHandler | DROP TABLE ... | `DELETE /<idx>` |
| AlterHandler | ALTER INDEX ... SETTINGS ... / ALTER TABLE ... RENAME TO ... | `PUT /<idx>/_settings` / `POST /_aliases` |

**可复用函数：**
- `query_helper.go` — `fetchDocByID()`、`buildResultFromSource()`，文档回查辅助
- `insert_handler.go` — `parseValue()`、`splitValues()`，值解析（被 UpdateHandler 复用）
- `update_handler.go` — `parseWhereClause()`，WHERE 解析（被 DeleteHandler 复用）
- `create_handler.go` — `parseSettings()`，SETTINGS 解析（被 AlterHandler 复用）
