# formatter 模块

> `internal/formatter/`

输出格式化，分为两个子包：`repl`（REPL 交互式）和 `cmd`（exec 非交互式）。

## 子包结构

```
internal/formatter/
├── repl/          REPL 交互式格式化
│   ├── formatter.go   Formatter 接口 + New() 工厂
│   ├── table.go       TableFormatter（基于 go-pretty）
│   └── json.go        JsonFormatter（ES 接口路径 + 响应体）
└── cmd/           exec 非交互式格式化
    ├── formatter.go   Format() 入口 + Options 结构体
    ├── json.go        JSON 格式化 + JSONPath 支持
    ├── csv.go         CSV 格式化 + 列过滤
    ├── yaml.go        YAML 格式化
    └── gotemplate.go  Go template 格式化
```

## repl 子包

REPL 模式下的格式化器，通过 `Formatter` 接口 + `New(format)` 工厂函数创建。

| 格式 | 实现 | 说明 |
|------|------|------|
| `table` | `TableFormatter` | 默认，含统计信息输出 |
| `json` | `JsonFormatter` | 输出 ES 接口路径 + 原始响应体（缩进格式化） |

## cmd 子包

exec 命令的格式化器，通过 `Format(result, w, opts)` 函数调用。

| 格式 | 说明 |
|------|------|
| `json` | 缩进输出，支持 `--jsonpath` 提取（基于 `github.com/PaesslerAG/jsonpath`） |
| `csv` | 标准 CSV 输出，支持 `--field` 列过滤 |
| `yaml` | YAML 输出（基于 `gopkg.in/yaml.v3`） |
| `go-template` | Go template 输出，数据上下文为 `*types.Result` |
