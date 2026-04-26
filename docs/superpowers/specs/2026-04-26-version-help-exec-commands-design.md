# 设计方案：新增 version / help / exec 命令

## Context

es-cli 当前只有一个 rootCmd，直接进入 REPL 交互模式。用户希望增加非交互式使用能力：通过 `exec` 子命令执行单条 SQL 并以多种格式输出结果，同时补充 `version` 和 `help` 命令。这为脚本化、管道化使用 es-cli 打下基础。

## 整体架构变更

```
es-cli                        → 进入 REPL（保持现有行为不变）
es-cli version                → 打印版本信息
es-cli help                   → cobra 自带帮助（无需额外实现）
es-cli exec -c 'SQL' [flags]  → 非交互式执行 SQL 并输出
```

rootCmd 的 PersistentFlags（--address/--username/--password/--password-stdin/--ca-cert）被所有子命令复用。

---

## 1. version 命令

### 1.1 新建 `internal/version/version.go`

定义版本信息变量，编译时通过 ldflags 注入：

```go
package version

var (
    Version   = "dev"
    GitCommit = "unknown"
    BuildDate = "unknown"
    GoVersion = runtime.Version()
)

func Print(w io.Writer) {
    fmt.Fprintf(w, "es-cli 版本: %s\n", Version)
    fmt.Fprintf(w, "Git 提交: %s\n", GitCommit)
    fmt.Fprintf(w, "构建日期: %s\n", BuildDate)
    fmt.Fprintf(w, "Go 版本: %s\n", GoVersion)
}
```

### 1.2 新建 `internal/cmd/version.go`

注册 versionCmd 子命令：

```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "打印版本信息",
    Run: func(cmd *cobra.Command, args []string) {
        version.Print(os.Stdout)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

### 1.3 更新构建命令

CLAUDE.md 中构建命令增加 ldflags 示例：

```bash
go build -ldflags "-X github.com/zhanghuangbin/es-cli/internal/version.Version=v0.1.0 \
  -X github.com/zhanghuangbin/es-cli/internal/version.GitCommit=$(git rev-parse --short HEAD) \
  -X github.com/zhanghuangbin/es-cli/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o ./_output/es-cli.exe ./cmd/es-cli/
```

---

## 2. help 命令

Cobra 自带 `help` 子命令，无需额外实现。仅需确保 rootCmd 的 `Short`/`Long` 描述准确即可。当前描述已经足够，无需修改。

---

## 3. exec 命令

### 3.1 命令定义 — `internal/cmd/exec.go`

**Flags：**

| Flag | 短标记 | 类型 | 默认值 | 说明 |
|------|--------|------|--------|------|
| `--command` | `-c` | string | 必填 | 要执行的 SQL 语句 |
| `--format` | `-f` | string | `json` | 输出格式: json\|csv\|yaml\|go-template |
| `--jsonpath` | | string | | JSONPath 表达式（仅 json 格式有效） |
| `--template` | | string | | Go template 模板字符串（仅 go-template 格式有效） |
| `--field` | | string | | 输出字段列表，逗号分隔（仅 csv 格式有效） |

**执行流程：**

1. 从 PersistentFlags 获取 ES 连接参数，创建 ES client
2. 解析 `-c` 中的 SQL，通过现有 Executor 执行（复用 handler 层）
3. 根据 `-f` 选择 exec 独立的 formatter 输出结果到 stdout
4. 执行完毕后进程退出（非交互）

**连接逻辑复用：** 当前 rootCmd.RunE 中有创建 client + ping 的逻辑，需要提取为公共函数供 exec 复用。新建 `internal/cmd/client.go`，提取 `newESClient()` 函数。

### 3.2 formatter 目录重构

将现有 `internal/formatter/` 拆分为两个子包，按用途隔离：

```
internal/formatter/
  repl/                  ← REPL 交互式格式化（原有代码移入）
    formatter.go         ← Formatter 接口 + New()
    table.go             ← 原 internal/formatter/table.go
    json.go              ← 原 internal/formatter/json.go
  cmd/                   ← exec 命令的非交互式格式化（新建）
    formatter.go         ← execFormat() + execFormatOptions
    json.go              ← json 格式 + jsonpath 支持
    csv.go               ← csv 格式 + --field 支持
    yaml.go              ← yaml 格式
    gotemplate.go        ← go-template 格式
```

**影响的现有引用：** `internal/executor/executor.go`、`internal/repl/repl.go`、`internal/cmd/root.go` 中的 `import "internal/formatter"` 需改为 `import "internal/formatter/repl"`。

### 3.3 exec formatter 实现 — `internal/formatter/cmd/`

在 `internal/formatter/cmd/formatter.go` 中定义入口函数：

```go
type Options struct {
    Format   string
    JSONPath string
    Template string
    Fields   []string
}

func Format(result *types.Result, w io.Writer, opts Options) error {
    switch opts.Format {
    case "json":
        return formatJSON(result, w, opts.JSONPath)
    case "csv":
        return formatCSV(result, w, opts.Fields)
    case "yaml":
        return formatYAML(result, w)
    case "go-template":
        return formatGoTemplate(result, w, opts.Template)
    default:
        return fmt.Errorf("未知格式: %s", opts.Format)
    }
}
```

### 3.4 各格式实现细节

#### json 格式

- 默认输出：将 Result 的 ES 原始响应体（`result.Source`）格式化缩进输出
- `--jsonpath` 模式：参考 kubectl 的 jsonpath 实现，使用第三方库或自行实现轻量级 JSONPath 解析
  - 推荐方案：使用标准库 `encoding/json` 解析为 `map[string]any`，再用轻量级 JSONPath 库（如 `github.com/PaesslerAG/jsonpath`）提取
  - 支持的语法示例：`--jsonpath '{.rows[0][0]}'`、`--jsonpath '{.columns[*].name}'`

#### csv 格式

- 使用标准库 `encoding/csv`
- 默认输出所有列（Result.Columns + Result.Rows）
- `--field field1,field2`：仅输出指定列，按 Columns 名称匹配
- 首行为表头

#### yaml 格式

- 将 Result 的 Columns + Rows 转为结构化数据后输出 YAML
- 使用第三方库 `gopkg.in/yaml.v3`
- 输出格式：列表形式，每行为一个 map（列名 → 值）

#### go-template 格式

- 参考 kubectl 的 `--template` / `-o go-template` 实现
- 使用标准库 `text/template`
- 模板的数据上下文为 `types.Result` 结构体，用户可以访问 `.Columns`、`.Rows`、`.Meta` 等字段
- 示例：`--template '{{range .Rows}}{{index . 0}}\t{{index . 1}}{{"\n"}}{{end}}'`

### 3.5 错误处理

- `-c` 未提供时报错并打印 usage
- `--format` 值不在支持范围时报错
- `--jsonpath` 仅在 `-f json` 时有效，否则报错提示
- `--template` 仅在 `-f go-template` 时有效，否则报错提示
- `--field` 仅在 `-f csv` 时有效，否则报错提示
- 指定的 field 名不存在于 Result.Columns 中时报错

---

## 4. 需要修改/新建的文件清单

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新建 | `internal/version/version.go` | 版本信息变量和打印函数 |
| 新建 | `internal/cmd/version.go` | version 子命令注册 |
| 新建 | `internal/cmd/exec.go` | exec 子命令定义和执行逻辑 |
| 新建 | `internal/cmd/client.go` | 提取 ES client 创建逻辑为公共函数 |
| 新建 | `internal/formatter/repl/formatter.go` | Formatter 接口 + New()（从原 formatter.go 移入） |
| 移动 | `internal/formatter/repl/table.go` | 原 internal/formatter/table.go |
| 移动 | `internal/formatter/repl/json.go` | 原 internal/formatter/json.go |
| 删除 | `internal/formatter/formatter.go` | 已移入 repl 子包 |
| 删除 | `internal/formatter/table.go` | 已移入 repl 子包 |
| 删除 | `internal/formatter/json.go` | 已移入 repl 子包 |
| 新建 | `internal/formatter/cmd/formatter.go` | exec 格式化入口（Format + Options） |
| 新建 | `internal/formatter/cmd/json.go` | json 格式 + jsonpath |
| 新建 | `internal/formatter/cmd/csv.go` | csv 格式 + --field |
| 新建 | `internal/formatter/cmd/yaml.go` | yaml 格式 |
| 新建 | `internal/formatter/cmd/gotemplate.go` | go-template 格式 |
| 修改 | `internal/executor/executor.go` | import 路径改为 `formatter/repl` |
| 修改 | `internal/repl/repl.go` | import 路径改为 `formatter/repl` |
| 修改 | `internal/cmd/root.go` | import 路径改为 `formatter/repl`；调用公共 client 函数 |
| 修改 | `go.mod` | 新增 `gopkg.in/yaml.v3` 依赖 |
| 修改 | `CLAUDE.md` | 更新构建命令，增加 ldflags 示例 |
| 更新 | `docs/references/cmd.md` | 更新命令模块文档 |
| 更新 | `docs/references/formatter.md` | 更新 formatter 模块文档 |

## 5. 新增依赖

- `gopkg.in/yaml.v3` — yaml 格式输出
- jsonpath：优先使用轻量级方案（标准库 `encoding/json` 解析 + 简单路径匹配），如不够再引入第三方库

## 6. 验证方案

```bash
# 编译
go build -ldflags "-X github.com/zhanghuangbin/es-cli/internal/version.Version=v0.1.0" -o ./_output/es-cli.exe ./cmd/es-cli/

# 验证 version
./_output/es-cli.exe version

# 验证 help
./_output/es-cli.exe help
./_output/es-cli.exe exec --help

# 验证 exec（需要本地 ES）
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SHOW TABLES' -f json
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SELECT * FROM my_index LIMIT 5' -f csv
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SELECT * FROM my_index LIMIT 5' -f csv --field name,age
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SHOW TABLES' -f yaml
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SHOW TABLES' -f json --jsonpath '{.rows[*][0]}'
./_output/es-cli.exe --address http://localhost:9200 exec -c 'SELECT * FROM my_index LIMIT 5' -f go-template --template '{{range .Rows}}{{index . 0}}{{"\n"}}{{end}}'

# 验证 REPL 仍正常
./_output/es-cli.exe --address http://localhost:9200

# 编译检查
go build ./...

# 运行测试
go test ./...
```
