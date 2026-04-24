# es-cli Design Spec

## Context

需要一个基于 Go 的 Elasticsearch CLI 工具，用户通过 REPL 模式使用 SQL 语法操作 Elasticsearch。作为练习项目，MVP 阶段先利用 ES 内置 SQL API 快速跑通，架构上预留扩展点，后续自行实现 SQL 解析器（SQL → ES DSL）。

## 技术栈

| 模块 | 库 | 版本/备注 |
|---|---|---|
| Go | 1.22+ | 支持泛型 |
| REPL/交互 | `github.com/c-bata/go-prompt` | 自动补全、语法高亮、历史记录 |
| ES 客户端 | `github.com/elastic/go-elasticsearch/v8` | 官方库，支持 ES 7.x/8.x |
| SQL 解析（未来） | `github.com/xwb1989/sqlparser` | vitess 系，AST 完整 |
| 表格输出 | `github.com/jedib0t/go-pretty/v6` | 多样式、分页 |
| CLI 框架 | `github.com/spf13/cobra` + `github.com/spf13/viper` | 命令行参数、配置管理 |

## 项目结构

```
es-cli/
├── cmd/
│   └── es-cli/
│       └── main.go                # 程序入口，初始化 cobra root command
├── internal/
│   ├── cmd/                       # cobra 命令定义
│   │   └── root.go                # root command, 连接参数 flags
│   ├── repl/                      # REPL 交互层
│   │   ├── repl.go                # REPL 主循环
│   │   ├── completer.go           # 自动补全（索引名、字段名、SQL 关键字）
│   │   ├── highlighter.go         # SQL 语法高亮
│   │   └── history.go             # 命令历史（持久化到文件）
│   ├── executor/                  # SQL 执行调度
│   │   └── executor.go            # 接收 SQL，调用 translator，返回结果
│   ├── translator/                # SQL 翻译层（核心扩展点）
│   │   ├── translator.go          # Translator 接口定义
│   │   └── builtin.go             # MVP 实现：调用 ES _sql API
│   └── formatter/                 # 输出格式化
│       ├── formatter.go           # Formatter 接口
│       ├── table.go               # 表格输出 (go-pretty)
│       ├── json.go                # JSON 输出
│       └── csv.go                 # CSV 输出
├── pkg/
│   └── es/                        # ES 客户端封装
│       ├── client.go              # 客户端创建、连接配置
│       └── config.go              # 连接参数结构体
├── go.mod
└── go.sum
```

### 分层原则

- **`cmd/es-cli/`** — 尽量薄，只做入口初始化
- **`internal/`** — 所有业务逻辑，不可被外部项目导入
- **`pkg/`** — ES 客户端封装，通用能力，可复用

## 核心接口设计

### Translator 接口（核心扩展点）

```go
// internal/translator/translator.go
type Result struct {
	Meta    Meta
    Columns []string
    Rows    [][]any
}

type Translator interface {
    // Translate 接收 SQL，返回查询结果
    // MVP 阶段直接调用 ES SQL API
    // 未来替换为：解析 SQL → 生成 DSL → 执行 → 返回结果
    Execute(ctx context.Context, sql string) (*Result, error)
}
```

### Formatter 接口

```go
// internal/formatter/formatter.go
type Formatter interface {
    Format(result *translator.Result, w io.Writer) error
}
```

### ES Client 封装

```go
// pkg/es/config.go
type Config struct {
    Addresses []string
    Username  string
    Password  string
    CACert    string  // TLS CA 证书路径
}

// pkg/es/client.go
func NewClient(cfg Config) (*elasticsearch.Client, error)
```

## 数据流

```
用户输入 SQL
    │
    ▼
  REPL (go-prompt)
    │ 补全/高亮/历史
    ▼
  Executor
    │
    ▼
  Translator (interface)
    │
    ├── [MVP] BuiltinTranslator → ES _sql API → 结果
    │
    └── [未来] CustomTranslator → SQL Parser → ES DSL → ES Search API → 结果
    │
    ▼
  Formatter (table/json/csv)
    │
    ▼
  终端输出
```

## MVP 功能范围

### 必须实现

1. **连接管理** — 通过命令行参数指定 ES 地址、用户名/密码、TLS 配置
2. **REPL 循环** — 进入交互式 SQL 输入环境
3. **自动补全** — SQL 关键字补全，索引名/字段名动态补全
4. **语法高亮** — SQL 关键字着色
5. **历史记录** — 上下箭头浏览，跨会话持久化到 `~/.es-cli/history`
6. **SQL 执行** — 通过 ES `_sql` API 执行 SQL（SELECT/INSERT/UPDATE/DELETE/DDL）
7. **表格输出** — 查询结果以格式化表格展示
8. **多格式输出** — 支持 table/json/csv 切换

### 内置命令（非 SQL）

| 命令 | 功能 |
|---|---|
| `.help` | 显示帮助信息 |
| `.format <type>` | 切换输出格式 (table/json/csv) |
| `.indices` | 列出所有索引 |
| `.schema <index>` | 显示索引 mapping |
| `.exit` / `Ctrl+D` | 退出 |

### 不在 MVP 范围

- 自定义 SQL 解析器（架构预留，后续实现）
- 多行 SQL 输入
- 查询结果分页滚动
- 连接配置文件 (~/.es-cli/config.yaml)
- 查询耗时统计

## 扩展计划

### Phase 2: 自定义 SQL 解析器

- 引入 `github.com/xwb1989/sqlparser` 解析 SQL AST
- 实现 `CustomTranslator`：SQL AST → ES DSL JSON
- 逐步覆盖 SELECT/WHERE/GROUP BY/ORDER BY/LIMIT/聚合函数
- 通过配置或命令切换 Builtin/Custom translator

### Phase 3: 高级特性

- 多行 SQL 输入（以 `;` 结尾触发执行）
- 查询耗时统计与 EXPLAIN 支持
- 连接配置文件持久化
- 查询结果导出到文件

## 验证方案

1. 启动本地 ES 实例（Docker）
2. 运行 `es-cli --address http://localhost:9200`
3. 验证进入 REPL 后：
   - Tab 键触发补全
   - 输入 `SELECT * FROM <index> LIMIT 10` 返回表格结果
   - `.indices` 列出索引
   - `.format json` 切换输出格式
   - 上下箭头浏览历史
   - `.exit` 正常退出
