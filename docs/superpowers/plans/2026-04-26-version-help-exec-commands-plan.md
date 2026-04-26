# 实施计划：新增 version / help / exec 命令

## Context

es-cli 当前只有 rootCmd（进入 REPL），需要新增非交互式能力：`version` 打印版本、`exec` 执行单条 SQL 并格式化输出（json/csv/yaml/go-template）。`help` 由 cobra 自带，无需额外实现。这为脚本化、管道化使用打下基础。

设计规格来自 `docs/superpowers/specs/2026-04-26-version-help-exec-commands-design.md`。

---

## 步骤 1：version 命令

### 1.1 新建 `internal/version/version.go`
- 定义 `Version`、`GitCommit`、`BuildDate` 变量（编译时 ldflags 注入），`GoVersion` 取 `runtime.Version()`
- `Print(w io.Writer)` 函数输出版本信息

### 1.2 新建 `internal/cmd/version.go`
- 定义 `versionCmd`，Use="version"，Short="打印版本信息"
- `init()` 中 `rootCmd.AddCommand(versionCmd)`

**验证**：`go build ./...` 通过

---

## 步骤 2：提取 ES client 创建逻辑

### 2.1 新建 `internal/cmd/client.go`
- 提取 `newESClient()` 函数：读取 passwordStdin、创建 `es.Config`、调用 `es.NewClient`、ping
- 返回 `(*elasticsearch.Client, error)`

### 2.2 修改 `internal/cmd/root.go`
- `RunE` 中改为调用 `newESClient()`，删除内联的 client 创建逻辑
- 移除 `root.go` 中不再需要的 import（`bufio`、`strings`）

**验证**：`go build ./...` 通过

---

## 步骤 3：formatter 目录重构

将 `internal/formatter/` 拆分为 `internal/formatter/repl/`（REPL 交互式）和 `internal/formatter/cmd/`（exec 非交互式）。

### 3.1 新建 `internal/formatter/repl/` 子包
- `internal/formatter/repl/formatter.go` — 移入 `Formatter` 接口 + `New()` 工厂函数，package 改为 `repl`
- `internal/formatter/repl/table.go` — 移入 `TableFormatter`，package 改为 `repl`
- `internal/formatter/repl/json.go` — 移入 `JsonFormatter`，package 改为 `repl`

### 3.2 删除旧文件
- 删除 `internal/formatter/formatter.go`
- 删除 `internal/formatter/table.go`
- 删除 `internal/formatter/json.go`

### 3.3 更新 import 引用（3 个文件）

| 文件 | 旧 import | 新 import |
|------|-----------|-----------|
| `internal/executor/executor.go` | `"...internal/formatter"` | `"...internal/formatter/repl"` |
| `internal/repl/repl.go` | `"...internal/formatter"` | `"...internal/formatter/repl"` |
| `internal/cmd/root.go` | `"...internal/formatter"` | `"...internal/formatter/repl"` |

注意：`executor.go` 中类型引用 `formatter.Formatter` → `repl.Formatter`，`formatter.New()` → `repl.New()`。`repl.go` 同理。

**验证**：`go build ./...` 通过

---

## 步骤 4：exec formatter 实现

### 4.1 新建 `internal/formatter/cmd/formatter.go`
- 定义 `Options` 结构体（Format, JSONPath, Template, Fields）
- 定义 `Format(result *types.Result, w io.Writer, opts Options) error`，switch 分发

### 4.2 新建 `internal/formatter/cmd/json.go`
- `formatJSON(result, w, jsonpath)` — 默认缩进输出 `result.Source`
- jsonpath 模式：用 `encoding/json` 解析为 `map[string]any`，使用轻量级第三方库 `github.com/PaesslerAG/jsonpath` 提取

### 4.3 新建 `internal/formatter/cmd/csv.go`
- `formatCSV(result, w, fields)` — 用 `encoding/csv`，首行表头，支持 `--field` 过滤列

### 4.4 新建 `internal/formatter/cmd/yaml.go`
- `formatYAML(result, w)` — 将 Columns+Rows 转为 `[]map[string]any` 后输出 YAML
- 依赖 `gopkg.in/yaml.v3`

### 4.5 新建 `internal/formatter/cmd/gotemplate.go`
- `formatGoTemplate(result, w, tmpl)` — 用 `text/template`，数据上下文为 `*types.Result`

**验证**：`go build ./...` 通过

---

## 步骤 5：exec 命令

### 5.1 新建 `internal/cmd/exec.go`
- 定义 `execCmd`，Use="exec"，Short="执行 SQL 并输出结果"
- Flags：`-c/--command`（必填）、`-f/--format`（默认 json）、`--jsonpath`、`--template`、`--field`
- RunE 流程：
  1. 调用 `newESClient()` 创建 client
  2. 解析 SQL 类型（`types.DetectSQLType`），从 handler map 获取对应 Handler 执行
  3. 调用 `cmdfmt.Format()` 输出结果到 stdout
- 错误校验：`-c` 未提供报错；`--jsonpath` 仅 json 格式有效；`--template` 仅 go-template 有效；`--field` 仅 csv 有效
- `init()` 中 `rootCmd.AddCommand(execCmd)`

**验证**：`go build ./...` 通过

---

## 步骤 6：依赖 & 文档更新

### 6.1 新增依赖
- `go get gopkg.in/yaml.v3`
- `go get github.com/PaesslerAG/jsonpath`
- `go mod tidy`

### 6.2 更新 `CLAUDE.md`
- 构建命令增加 ldflags 示例

### 6.3 更新 `docs/references/cmd.md`
- 补充 version、exec 子命令说明

### 6.4 更新 `docs/references/formatter.md`
- 补充 repl/cmd 子包结构和 exec formatter 说明

**验证**：`go build ./...` 和 `go test ./...` 通过

---

## 关键文件清单

| 操作 | 文件路径 |
|------|----------|
| 新建 | `internal/version/version.go` |
| 新建 | `internal/cmd/version.go` |
| 新建 | `internal/cmd/client.go` |
| 新建 | `internal/cmd/exec.go` |
| 新建 | `internal/formatter/repl/formatter.go` |
| 新建 | `internal/formatter/repl/table.go` |
| 新建 | `internal/formatter/repl/json.go` |
| 删除 | `internal/formatter/formatter.go` |
| 删除 | `internal/formatter/table.go` |
| 删除 | `internal/formatter/json.go` |
| 新建 | `internal/formatter/cmd/formatter.go` |
| 新建 | `internal/formatter/cmd/json.go` |
| 新建 | `internal/formatter/cmd/csv.go` |
| 新建 | `internal/formatter/cmd/yaml.go` |
| 新建 | `internal/formatter/cmd/gotemplate.go` |
| 修改 | `internal/executor/executor.go`（import 路径） |
| 修改 | `internal/repl/repl.go`（import 路径） |
| 修改 | `internal/cmd/root.go`（import 路径 + 调用 newESClient） |
| 修改 | `go.mod`（新增依赖） |
| 修改 | `CLAUDE.md` |
| 修改 | `docs/references/cmd.md` |
| 修改 | `docs/references/formatter.md` |

## 复用的现有模块

- `pkg/es/client.go` — `es.NewClient(cfg)` + `es.Config` 创建 ES 客户端
- `internal/handler/` — 所有 Handler 实现（QueryHandler, InsertHandler 等），exec 命令复用
- `internal/types/result.go` — `Result` 结构体（Columns, Rows, Source, Meta）
- `internal/types/sqltype.go` — `DetectSQLType()` 解析 SQL 类型

## 验证方案

```bash
# 编译检查
go build ./...

# 带 ldflags 编译
go build -ldflags "-X github.com/zhanghuangbin/es-cli/internal/version.Version=v0.1.0 \
  -X github.com/zhanghuangbin/es-cli/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X github.com/zhanghuangbin/es-cli/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o ./_output/es-cli.exe ./cmd/es-cli/

# 验证 version
./_output/es-cli.exe version

# 验证 help
./_output/es-cli.exe help
./_output/es-cli.exe exec --help

# 运行测试
go test ./...

# 验证 exec（需要本地 ES）
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SHOW TABLES' -f json
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SHOW TABLES' -f csv
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SHOW TABLES' -f yaml
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SHOW TABLES' -f json --jsonpath '{.rows[*][0]}'
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SHOW TABLES' -f go-template --template '{{range .Rows}}{{index . 0}}{{"\n"}}{{end}}'
```
