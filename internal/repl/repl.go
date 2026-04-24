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
