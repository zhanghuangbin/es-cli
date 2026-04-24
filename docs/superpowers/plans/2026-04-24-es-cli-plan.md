# es-cli 实现计划

> **给 Agent 执行者：** 必须使用 superpowers:subagent-driven-development 技能来逐任务执行此计划。步骤使用 `- [ ]` 复选框语法跟踪进度。

**目标：** 构建一个基于 Go 的 REPL 工具，让用户通过 SQL 语法查询和管理 Elasticsearch。

**架构：** 分层设计：REPL → Executor → Translator → ES Client。Translator 是接口——MVP 使用 ES 内置 `_sql` API，后续阶段替换为自定义 SQL 解析器。输出通过 Formatter 接口，MVP 仅实现 table 格式（JSON/CSV 后续扩展）。

**技术栈：** Go 1.22+、go-prompt（REPL）、go-elasticsearch/v8（ES 客户端）、go-pretty/v6（表格）、cobra+viper（CLI）

---

## 文件清单

| 文件 | 职责                                               |
|---|--------------------------------------------------|
| `cmd/es-cli/main.go` | 程序入口，调用 root command                             |
| `internal/cmd/root.go` | Cobra root command，连接参数 flags，启动 REPL            |
| `pkg/es/config.go` | ES 连接参数 `Config` 结构体                             |
| `pkg/es/client.go` | `NewClient()` 工厂函数，创建 `*elasticsearch.Client`    |
| `internal/translator/translator.go` | `Translator` 接口、`Result` 结构体、`Meta` 结构体          |
| `internal/translator/builtin.go` | `BuiltinTranslator` — 调用 ES `_sql` API           |
| `internal/formatter/formatter.go` | `Formatter` 接口                                   |
| `internal/formatter/table.go` | 表格格式化器，使用 go-pretty                              |
| `internal/executor/executor.go` | `Executor` — 将 SQL 分发到 translator，结果传给 formatter |
| `internal/repl/repl.go` | REPL 主循环，使用 go-prompt                            |
| `internal/repl/completer.go` | 自动补全：SQL 关键字、索引名、字段名                             |
| `internal/repl/history.go` | 持久化命令历史到 `~/.es-cli/history`                     |

---

### 任务 1：项目初始化与 CLI 骨架

**文件：**
- 创建：`go.mod`
- 创建：`cmd/es-cli/main.go`
- 创建：`internal/cmd/root.go`

- [ ] **步骤 1：初始化 Go 模块**

```bash
cd D:/data/goproject/es-cli
go mod init github.com/zhanghuangbin/es-cli
```

- [ ] **步骤 2：安装核心依赖**

```bash
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/elastic/go-elasticsearch/v8@latest
go get github.com/c-bata/go-prompt@latest
go get github.com/jedib0t/go-pretty/v6@latest
```

- [ ] **步骤 3：创建目录结构**

```bash
mkdir -p cmd/es-cli
mkdir -p internal/cmd
mkdir -p internal/repl
mkdir -p internal/executor
mkdir -p internal/translator
mkdir -p internal/formatter
mkdir -p pkg/es
```

- [ ] **步骤 4：编写 `internal/cmd/root.go`**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	addresses []string
	username  string
	password  string
	caCert    string
)

var rootCmd = &cobra.Command{
	Use:   "es-cli",
	Short: "基于 SQL 的 Elasticsearch CLI",
	Long:  "一个交互式 REPL 工具，让你通过 SQL 语法查询和管理 Elasticsearch。",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("es-cli 已连接（REPL 尚未实现）")
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&addresses, "address", []string{"http://localhost:9200"}, "Elasticsearch 地址")
	rootCmd.PersistentFlags().StringVar(&username, "username", "", "Elasticsearch 用户名")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Elasticsearch 密码")
	rootCmd.PersistentFlags().StringVar(&caCert, "ca-cert", "", "TLS CA 证书路径")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **步骤 5：编写 `cmd/es-cli/main.go`**

```go
package main

import (
	"github.com/zhanghuangbin/es-cli/internal/cmd"
)

func main() {
	cmd.Execute()
}
```

- [ ] **步骤 6：运行验证骨架是否正常**

```bash
go run cmd/es-cli/main.go --help
```

预期输出：显示 `es-cli` 描述和 flags（`--address`、`--username`、`--password`、`--ca-cert`）。

- [ ] **步骤 7：提交**

```bash
git init
git add .
git commit -m "feat: 项目初始化，cobra CLI 骨架"
```

---

### 任务 2：ES 客户端层

**文件：**
- 创建：`pkg/es/config.go`
- 创建：`pkg/es/client.go`

- [ ] **步骤 1：编写 `pkg/es/config.go`**

```go
package es

type Config struct {
	Addresses []string
	Username  string
	Password  string
	CACert    string
}
```

- [ ] **步骤 2：编写 `pkg/es/client.go`**

```go
package es

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewClient(cfg Config) (*elasticsearch.Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	if cfg.CACert != "" {
		caCert, err := os.ReadFile(cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("读取 CA 证书失败: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("解析 CA 证书失败")
		}
		esCfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 ES 客户端失败: %w", err)
	}
	return client, nil
}
```

- [ ] **步骤 3：将 ES 客户端接入 root command**

修改 `internal/cmd/root.go` — 更新 `RunE` 以创建并 ping ES 客户端：

```go
RunE: func(cmd *cobra.Command, args []string) error {
	client, err := es.NewClient(es.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
		CACert:    caCert,
	})
	if err != nil {
		return fmt.Errorf("连接 ES 失败: %w", err)
	}
	res, err := client.Ping()
	if err != nil {
		return fmt.Errorf("ping ES 失败: %w", err)
	}
	defer res.Body.Close()
	fmt.Printf("已连接到 Elasticsearch (%s)\n", addresses)
	return nil
},
```

添加 import：`"github.com/zhanghuangbin/es-cli/pkg/es"`

- [ ] **步骤 4：编译验证**

```bash
go build ./...
```

预期：编译无错误。

- [ ] **步骤 5：提交**

```bash
git add .
git commit -m "feat: 添加 ES 客户端层，支持 TLS"
```

---

### 任务 3：Translator 接口与 BuiltinTranslator

**文件：**
- 创建：`internal/translator/translator.go`
- 创建：`internal/translator/builtin.go`

- [ ] **步骤 1：编写 `internal/translator/translator.go`**

```go
package translator

import "context"

type Meta struct {
	Status  int
	Message string
}

type Result struct {
	Meta    Meta
	Columns []string
	Rows    [][]any
}

type Translator interface {
	Execute(ctx context.Context, sql string) (*Result, error)
}
```

- [ ] **步骤 2：编写 `internal/translator/builtin.go`**

```go
package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
)

type BuiltinTranslator struct {
	client *elasticsearch.Client
}

func NewBuiltinTranslator(client *elasticsearch.Client) *BuiltinTranslator {
	return &BuiltinTranslator{client: client}
}

type sqlRequest struct {
	Query string `json:"query"`
}

type sqlResponse struct {
	Columns []sqlColumn     `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

type sqlColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (t *BuiltinTranslator) Execute(ctx context.Context, sql string) (*Result, error) {
	body, err := json.Marshal(sqlRequest{Query: sql})
	if err != nil {
		return nil, fmt.Errorf("序列化 SQL 请求失败: %w", err)
	}

	res, err := t.client.SQL.Query(
		bytes.NewReader(body),
		t.client.SQL.Query.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("执行 SQL 失败: %w", err)
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if res.IsError() {
		return &Result{
			Meta: Meta{Status: res.StatusCode, Message: string(respBody)},
		}, fmt.Errorf("ES 错误 [%d]: %s", res.StatusCode, respBody)
	}

	var sqlRes sqlResponse
	if err := json.Unmarshal(respBody, &sqlRes); err != nil {
		return nil, fmt.Errorf("解析 SQL 响应失败: %w", err)
	}

	columns := make([]string, len(sqlRes.Columns))
	for i, col := range sqlRes.Columns {
		columns[i] = col.Name
	}

	rows := make([][]any, len(sqlRes.Rows))
	for i, row := range sqlRes.Rows {
		rows[i] = row
	}

	return &Result{
		Meta:    Meta{Status: res.StatusCode},
		Columns: columns,
		Rows:    rows,
	}, nil
}
```

- [ ] **步骤 3：编译验证**

```bash
go build ./...
```

预期：编译无错误。

- [ ] **步骤 4：提交**

```bash
git add .
git commit -m "feat: 添加 Translator 接口和 BuiltinTranslator（ES SQL API）"
```

---

### 任务 4：Formatter 接口与表格实现

**文件：**
- 创建：`internal/formatter/formatter.go`
- 创建：`internal/formatter/table.go`

- [ ] **步骤 1：编写 `internal/formatter/formatter.go`**

```go
package formatter

import (
	"fmt"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/translator"
)

type Formatter interface {
	Format(result *translator.Result, w io.Writer) error
}

func New(format string) (Formatter, error) {
	switch format {
	case "table":
		return &TableFormatter{}, nil
	case "json", "csv":
		return nil, fmt.Errorf("格式 '%s' 暂未实现，敬请期待", format)
	default:
		return nil, fmt.Errorf("未知格式: %s（可选: table, json, csv）", format)
	}
}
```

- [ ] **步骤 2：编写 `internal/formatter/table.go`**

```go
package formatter

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/zhanghuangbin/es-cli/internal/translator"
)

type TableFormatter struct{}

func (f *TableFormatter) Format(result *translator.Result, w io.Writer) error {
	if len(result.Columns) == 0 {
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(w)

	header := make(table.Row, len(result.Columns))
	for i, col := range result.Columns {
		header[i] = col
	}
	t.AppendHeader(header)

	for _, row := range result.Rows {
		tableRow := make(table.Row, len(row))
		for i, val := range row {
			tableRow[i] = val
		}
		t.AppendRow(tableRow)
	}

	t.SetStyle(table.StyleLight)
	t.Render()
	return nil
}
```

- [ ] **步骤 3：编译验证**

```bash
go build ./...
```

预期：编译无错误。

- [ ] **步骤 4：提交**

```bash
git add .
git commit -m "feat: 添加 Formatter 接口和表格格式化器"
```

---

### 任务 5：Executor 执行器

**文件：**
- 创建：`internal/executor/executor.go`

- [ ] **步骤 1：编写 `internal/executor/executor.go`**

```go
package executor

import (
	"context"
	"io"

	"github.com/zhanghuangbin/es-cli/internal/formatter"
	"github.com/zhanghuangbin/es-cli/internal/translator"
)

type Executor struct {
	translator translator.Translator
	formatter  formatter.Formatter
	output     io.Writer
}

func New(t translator.Translator, f formatter.Formatter, output io.Writer) *Executor {
	return &Executor{
		translator: t,
		formatter:  f,
		output:     output,
	}
}

func (e *Executor) SetFormatter(f formatter.Formatter) {
	e.formatter = f
}

func (e *Executor) Execute(sql string) error {
	result, err := e.translator.Execute(context.Background(), sql)
	if err != nil {
		return err
	}
	return e.formatter.Format(result, e.output)
}
```

- [ ] **步骤 2：编译验证**

```bash
go build ./...
```

预期：编译无错误。

- [ ] **步骤 3：提交**

```bash
git add .
git commit -m "feat: 添加 Executor 执行器"
```

---

### 任务 6：REPL — 主循环与内置命令

**文件：**
- 创建：`internal/repl/repl.go`

- [ ] **步骤 1：编写 `internal/repl/repl.go`**

```go
package repl

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/executor"
	"github.com/zhanghuangbin/es-cli/internal/formatter"
)

type REPL struct {
	executor  *executor.Executor
	client    *elasticsearch.Client
	format    string
	completer *Completer
	history   *History
}

func New(exec *executor.Executor, client *elasticsearch.Client) *REPL {
	return &REPL{
		executor:  exec,
		client:    client,
		format:    "table",
		completer: NewCompleter(client),
		history:   NewHistory(),
	}
}

func (r *REPL) Run() {
	fmt.Println("输入 SQL 查询 Elasticsearch。输入 .help 查看可用命令。")

	p := prompt.New(
		r.executeInput,
		r.completer.Complete,
		prompt.OptionTitle("es-cli"),
		prompt.OptionPrefix("es> "),
		prompt.OptionHistory(r.history.Entries()),
		prompt.OptionPrefixTextColor(prompt.Cyan),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSelectedSuggestionBGColor(prompt.Cyan),
		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionTextColor(prompt.LightGray),
	)
	p.Run()
}

func (r *REPL) executeInput(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	r.history.Add(input)

	if strings.HasPrefix(input, ".") {
		r.handleBuiltinCommand(input)
		return
	}

	if err := r.executor.Execute(input); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
	}
}

func (r *REPL) handleBuiltinCommand(input string) {
	parts := strings.Fields(input)
	cmd := parts[0]

	switch cmd {
	case ".help":
		r.showHelp()
	case ".format":
		if len(parts) < 2 {
			fmt.Printf("当前格式: %s\n", r.format)
			return
		}
		r.setFormat(parts[1])
	case ".indices":
		r.showIndices()
	case ".schema":
		if len(parts) < 2 {
			fmt.Println("用法: .schema <索引名>")
			return
		}
		r.showSchema(parts[1])
	case ".exit":
		fmt.Println("再见！")
		os.Exit(0)
	default:
		fmt.Printf("未知命令: %s。输入 .help 查看可用命令。\n", cmd)
	}
}

func (r *REPL) showHelp() {
	fmt.Println(`可用命令:
  .help            显示帮助信息
  .format <类型>   设置输出格式 (table, json*, csv*)
  .indices         列出所有索引
  .schema <索引名> 显示索引 mapping
  .exit            退出 es-cli
  Ctrl+D           退出 es-cli

  * json/csv 格式暂未实现

输入 SQL 语句查询 Elasticsearch。
示例: SELECT * FROM my_index LIMIT 10`)
}

func (r *REPL) setFormat(format string) {
	f, err := formatter.New(format)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	r.format = format
	r.executor.SetFormatter(f)
	fmt.Printf("输出格式已设置为: %s\n", format)
}

func (r *REPL) showIndices() {
	if err := r.executor.Execute("SHOW TABLES"); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
	}
}

func (r *REPL) showSchema(index string) {
	sql := fmt.Sprintf("DESCRIBE %s", index)
	if err := r.executor.Execute(sql); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
	}
}
```

- [ ] **步骤 2：编译验证**

```bash
go build ./...
```

预期：编译会失败，因为 `Completer` 和 `History` 尚未创建——这是预期的，继续任务 7 和 8。

- [ ] **步骤 3：提交**

```bash
git add .
git commit -m "feat: 添加 REPL 主循环和内置命令"
```

---

### 任务 7：REPL — 自动补全

**文件：**
- 创建：`internal/repl/completer.go`

- [ ] **步骤 1：编写 `internal/repl/completer.go`**

```go
package repl

import (
	"encoding/json"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/elastic/go-elasticsearch/v8"
)

var sqlKeywords = []prompt.Suggest{
	{Text: "SELECT", Description: "查询数据"},
	{Text: "FROM", Description: "指定表/索引"},
	{Text: "WHERE", Description: "过滤条件"},
	{Text: "AND", Description: "逻辑与"},
	{Text: "OR", Description: "逻辑或"},
	{Text: "NOT", Description: "逻辑非"},
	{Text: "IN", Description: "集合包含"},
	{Text: "LIKE", Description: "模式匹配"},
	{Text: "BETWEEN", Description: "范围"},
	{Text: "IS", Description: "IS NULL / IS NOT NULL"},
	{Text: "NULL", Description: "空值"},
	{Text: "ORDER BY", Description: "排序"},
	{Text: "GROUP BY", Description: "分组"},
	{Text: "HAVING", Description: "分组过滤"},
	{Text: "LIMIT", Description: "限制行数"},
	{Text: "AS", Description: "别名"},
	{Text: "COUNT", Description: "计数聚合"},
	{Text: "SUM", Description: "求和聚合"},
	{Text: "AVG", Description: "平均值聚合"},
	{Text: "MIN", Description: "最小值"},
	{Text: "MAX", Description: "最大值"},
	{Text: "INSERT INTO", Description: "插入数据"},
	{Text: "UPDATE", Description: "更新数据"},
	{Text: "DELETE FROM", Description: "删除数据"},
	{Text: "SET", Description: "设置值"},
	{Text: "VALUES", Description: "指定值"},
	{Text: "SHOW TABLES", Description: "列出索引"},
	{Text: "DESCRIBE", Description: "显示 mapping"},
}

var builtinCommands = []prompt.Suggest{
	{Text: ".help", Description: "显示帮助"},
	{Text: ".format", Description: "设置输出格式 (table/json/csv)"},
	{Text: ".indices", Description: "列出所有索引"},
	{Text: ".schema", Description: "显示索引 mapping"},
	{Text: ".exit", Description: "退出 es-cli"},
}

type Completer struct {
	client        *elasticsearch.Client
	cachedIndices []prompt.Suggest
}

func NewCompleter(client *elasticsearch.Client) *Completer {
	return &Completer{client: client}
}

func (c *Completer) Complete(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	word := d.GetWordBeforeCursor()

	if word == "" {
		return nil
	}

	if strings.HasPrefix(text, ".") {
		return prompt.FilterHasPrefix(builtinCommands, word, true)
	}

	upperText := strings.ToUpper(text)
	if strings.Contains(upperText, "FROM") || strings.Contains(upperText, "INTO") ||
		strings.Contains(upperText, "UPDATE") || strings.Contains(upperText, "DESCRIBE") ||
		strings.Contains(upperText, ".SCHEMA") {
		indices := c.getIndices()
		return prompt.FilterHasPrefix(indices, word, true)
	}

	return prompt.FilterHasPrefix(sqlKeywords, word, true)
}

func (c *Completer) getIndices() []prompt.Suggest {
	if c.cachedIndices != nil {
		return c.cachedIndices
	}

	res, err := c.client.Cat.Indices(
		c.client.Cat.Indices.WithFormat("json"),
		c.client.Cat.Indices.WithH("index"),
	)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil
	}

	var indices []struct {
		Index string `json:"index"`
	}
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil
	}

	suggestions := make([]prompt.Suggest, 0, len(indices))
	for _, idx := range indices {
		if !strings.HasPrefix(idx.Index, ".") {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        idx.Index,
				Description: "索引",
			})
		}
	}
	c.cachedIndices = suggestions
	return suggestions
}
```

- [ ] **步骤 2：编译验证**

```bash
go build ./...
```

预期：如果 History 尚未创建则仍会失败，继续任务 8。

- [ ] **步骤 3：提交**

```bash
git add .
git commit -m "feat: 添加 SQL 和索引自动补全"
```

---

### 任务 8：REPL — 历史记录

**文件：**
- 创建：`internal/repl/history.go`

- [ ] **步骤 1：编写 `internal/repl/history.go`**

```go
package repl

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type History struct {
	entries  []string
	filePath string
}

func NewHistory() *History {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".es-cli")
	os.MkdirAll(dir, 0755)

	h := &History{
		filePath: filepath.Join(dir, "history"),
	}
	h.load()
	return h
}

func (h *History) Entries() []string {
	return h.entries
}

func (h *History) Add(entry string) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return
	}
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == entry {
		return
	}
	h.entries = append(h.entries, entry)
	h.save(entry)
}

func (h *History) load() {
	f, err := os.Open(h.filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}
}

func (h *History) save(entry string) {
	f, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(entry + "\n")
}
```

- [ ] **步骤 2：编译验证**

```bash
go build ./...
```

预期：编译无错误，所有包都已就绪。

- [ ] **步骤 3：提交**

```bash
git add .
git commit -m "feat: 添加持久化命令历史"
```

---

### 任务 9：集成 — 串联所有层

**文件：**
- 修改：`internal/cmd/root.go`

- [ ] **步骤 1：更新 `internal/cmd/root.go`，串联所有层**

替换整个文件内容：

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhanghuangbin/es-cli/internal/executor"
	"github.com/zhanghuangbin/es-cli/internal/formatter"
	"github.com/zhanghuangbin/es-cli/internal/repl"
	"github.com/zhanghuangbin/es-cli/internal/translator"
	"github.com/zhanghuangbin/es-cli/pkg/es"
)

var (
	addresses []string
	username  string
	password  string
	caCert    string
)

var rootCmd = &cobra.Command{
	Use:   "es-cli",
	Short: "基于 SQL 的 Elasticsearch CLI",
	Long:  "一个交互式 REPL 工具，让你通过 SQL 语法查询和管理 Elasticsearch。",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := es.NewClient(es.Config{
			Addresses: addresses,
			Username:  username,
			Password:  password,
			CACert:    caCert,
		})
		if err != nil {
			return fmt.Errorf("连接 ES 失败: %w", err)
		}

		res, err := client.Ping()
		if err != nil {
			return fmt.Errorf("ping ES 失败: %w", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("ES ping 失败: %s", res.String())
		}

		fmt.Printf("已连接到 Elasticsearch (%s)\n", addresses)

		trans := translator.NewBuiltinTranslator(client)
		fmtr, _ := formatter.New("table")
		exec := executor.New(trans, fmtr, os.Stdout)

		r := repl.New(exec, client)
		r.Run()
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringSliceVar(&addresses, "address", []string{"http://localhost:9200"}, "Elasticsearch 地址")
	rootCmd.PersistentFlags().StringVar(&username, "username", "", "Elasticsearch 用户名")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Elasticsearch 密码")
	rootCmd.PersistentFlags().StringVar(&caCert, "ca-cert", "", "TLS CA 证书路径")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **步骤 2：运行 `go mod tidy`**

```bash
go mod tidy
```

- [ ] **步骤 3：构建最终二进制**

```bash
go build -o es-cli.exe ./cmd/es-cli/
```

预期：生成 `es-cli.exe` 二进制文件，无编译错误。

- [ ] **步骤 4：提交**

```bash
git add .
git commit -m "feat: 串联所有层 — MVP 完成"
```

---

### 任务 10：端到端验证

- [ ] **步骤 1：通过 Docker 启动 Elasticsearch**

```bash
docker run -d --name es-test -p 9200:9200 -e "discovery.type=single-node" -e "xpack.security.enabled=false" docker.elastic.co/elasticsearch/elasticsearch:8.13.0
```

等待 ES 就绪：

```bash
curl http://localhost:9200
```

预期：返回包含集群信息的 JSON 响应。

- [ ] **步骤 2：创建测试数据**

```bash
curl -X POST "http://localhost:9200/test_index/_doc" -H "Content-Type: application/json" -d "{\"name\":\"Alice\",\"age\":30,\"city\":\"Beijing\"}"
curl -X POST "http://localhost:9200/test_index/_doc" -H "Content-Type: application/json" -d "{\"name\":\"Bob\",\"age\":25,\"city\":\"Shanghai\"}"
curl -X POST "http://localhost:9200/test_index/_doc" -H "Content-Type: application/json" -d "{\"name\":\"Charlie\",\"age\":35,\"city\":\"Guangzhou\"}"
```

- [ ] **步骤 3：运行 es-cli 验证所有功能**

```bash
go run cmd/es-cli/main.go --address http://localhost:9200
```

在 REPL 中测试以下操作：

| 输入 | 预期结果 |
|---|---|
| `SELECT * FROM test_index LIMIT 10` | 显示 3 行数据的表格（Alice、Bob、Charlie） |
| 输入 `SEL` 后按 Tab | 自动补全为 `SELECT` |
| 输入 `FROM ` 后按 Tab | 显示 `test_index` 建议 |
| `.indices` | 列出 `test_index` |
| `.schema test_index` | 显示 mapping（name、age、city 字段） |
| `.format json` | 输出 "错误: 格式 'json' 暂未实现，敬请期待" |
| `.help` | 显示帮助信息 |
| 按上箭头 | 显示上一条命令 |
| `.exit` | 输出 "再见！" 并退出 |

- [ ] **步骤 4：验证历史记录持久化**

```bash
go run cmd/es-cli/main.go --address http://localhost:9200
```

按上箭头——应该能看到上一次会话的命令。

- [ ] **步骤 5：清理 Docker**

```bash
docker stop es-test && docker rm es-test
```

- [ ] **步骤 6：最终提交**

```bash
git add .
git commit -m "chore: MVP 端到端验证完成"
```
