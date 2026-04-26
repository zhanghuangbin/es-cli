# repl 模块

> `internal/repl/`

REPL 交互层，基于 `go-prompt`。

- `repl.go` — 主循环，处理输入分发（`.`开头→内置命令，`\`结尾→续行，其他→`executor.Execute()`）
- `completer.go` — 自动补全（SQL 关键字 + ES 索引名 + 内置命令），索引名从 `_cat/indices` 获取并缓存
- `history.go` — 命令历史持久化到 `~/.es-cli/history`，自动去重

内置命令：`.help`、`.ping`、`.format <type>`、`.indices`、`.desc <索引名>`、`.exit`
