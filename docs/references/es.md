# pkg/es 模块

> `pkg/es/`

ES 客户端封装，可被外部项目导入。

- `config.go` — `Config` 结构体（Addresses、Username、Password、CACert）
- `client.go` — `NewClient(cfg)` 创建 ES 客户端（支持 TLS）
- `rest.go` — `DoRequest(ctx, client, method, path, body)` 通用 HTTP 请求函数，所有 Handler（QueryHandler 除外）通过此函数与 ES 交互
