# es-cli

> **给 Agent 执行者：** 必须渐进式披露信息，按需阅读。

## 架构

分层设计，数据流为：**用户输入 → REPL → Executor → Translator → ES Client → Formatter → 终端输出**

```
cmd/es-cli/main.go          # 程序入口，仅调用 cmd.Execute()
internal/cmd/               # cobra的各种命令
internal/repl/              # REPL 交互层（go-prompt）
internal/executor/           # SQL 执行调度，接收 SQL 调用 Translator，结果传给 Formatter
internal/translator/         # SQL 执行的抽象层，将用户的查询请求，翻译成查询的语言
internal/formatter/          # 输出格式化
pkg/es/                      # ES 客户端封装（可被外部项目导入）
```

### 核心接口

- **`Translator`**（`internal/translator/translator.go`）：SQL 执行的抽象层。MVP 阶段只有 `BuiltinTranslator`（直接调用 ES SQL API）。后续扩展时实现 `CustomTranslator`（SQL Parser → ES DSL），无需修改上层代码。
- **`Formatter`**（`internal/formatter/formatter.go`）：输出格式化抽象。通过 `New(format)` 工厂函数创建，目前仅支持 `table`，`json` 和 `csv` 预留了接口但未实现。

### 分层原则

- `cmd/es-cli/` — 尽量薄，只做入口
- `internal/` — 所有业务逻辑，不可被外部导入
- `pkg/` — 通用能力，可复用

## 技术栈

| 模块 | 库 |
|---|---|
| CLI 框架 | `github.com/spf13/cobra` |
| REPL 交互 | `github.com/c-bata/go-prompt` |
| ES 客户端 | `github.com/elastic/go-elasticsearch/v8` |
| 表格输出 | `github.com/jedib0t/go-pretty/v6` |
| SQL 解析（未来） | `github.com/xwb1989/sqlparser`（计划中） |


## 阅读指南索引

### 核心模块

| 模块 | 指南 | 路径 | 说明 |
|------|------|------|------|