# cmd 模块

> `cmd/es-cli/main.go` + `internal/cmd/root.go`

程序入口。`main.go` 仅调用 `cmd.Execute()`，`root.go` 定义 CLI 参数（cobra），创建 ES 客户端、Formatter、Executor，启动 REPL。

CLI 参数：`--address`（ES 地址）、`--username`、`--password`、`--password-stdin`、`--ca-cert`
