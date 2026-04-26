# es-cli 模块总纲

> **给 Agent 执行者：** 本文件是项目的导航索引，按需阅读对应模块文档。

## 架构概览

数据流：**用户输入 → REPL → Executor → Handler → ES Client → Formatter → 终端输出**

```
cmd/es-cli/          程序入口
internal/cmd/        CLI 参数定义、启动流程
internal/repl/       REPL 交互层
internal/executor/   SQL 调度（按 SQLType 分发到 Handler）
internal/handler/    SQL 具体执行（每种 SQL 类型一个 Handler）
internal/types/      共享类型（SQLType 枚举、Result 结构体）
internal/formatter/  输出格式化（table、json）
pkg/es/              ES 客户端封装（可外部导入）
```

## 模块索引

| 模块 | 指南 | 一句话说明 |
|------|------|-----------|
| cmd | [cmd.md](cmd.md) | 入口 + CLI 参数（cobra） |
| repl | [repl.md](repl.md) | REPL 交互、补全、历史 |
| executor | [executor.md](executor.md) | SQL 调度器 |
| handler | [handler.md](handler.md) | 7 种 SQL Handler 实现 |
| types | [types.md](types.md) | SQLType、Result、Meta |
| formatter | [formatter.md](formatter.md) | 输出格式化（table/json） |
| es | [es.md](es.md) | ES 客户端 + 通用请求函数 |
