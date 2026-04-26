package repl

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/zhanghuangbin/es-cli/internal/executor"
	replFmt "github.com/zhanghuangbin/es-cli/internal/formatter/repl"
)

type REPL struct {
	executor        *executor.Executor
	client          *elasticsearch.Client
	format          string
	completer       *Completer
	history         *History
	multilineBuffer string
	addresses       []string
}

func New(exec *executor.Executor, client *elasticsearch.Client, addresses []string) *REPL {
	return &REPL{
		executor:  exec,
		client:    client,
		format:    "table",
		completer: NewCompleter(client),
		history:   NewHistory(),
		addresses: addresses,
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

	// 处理反斜杠续行
	if strings.HasSuffix(input, `\`) {
		line := strings.TrimSuffix(input, `\`)
		r.multilineBuffer += line + " "
		return
	}

	// 拼接缓冲区中的续行
	if r.multilineBuffer != "" {
		input = r.multilineBuffer + input
		r.multilineBuffer = ""
	}

	if strings.HasSuffix(input, ";") {
		input = strings.TrimSuffix(input, ";")
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
	case ".ping":
		r.handlePing()
	case ".format":
		if len(parts) < 2 {
			fmt.Printf("当前格式: %s\n", r.format)
			return
		}
		r.setFormat(parts[1])
	case ".indices":
		r.showIndices()
	case ".desc":
		if len(parts) < 2 {
			fmt.Println("用法: .desc <索引名>")
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
  .ping            测试 ES 连接是否正常
  .format <类型>   设置输出格式 (table, json*)
  .indices         列出所有索引
  .desc <索引名>    显示索引 mapping
  .exit            退出 es-cli
  Ctrl+D           退出 es-cli


输入 SQL 语句查询 Elasticsearch。
示例: SELECT * FROM my_index LIMIT 10

DDL/DML 示例:
  CREATE TABLE my_index (name TEXT, age INTEGER) SETTINGS (number_of_shards=3)
  DROP TABLE my_index
  INSERT INTO my_index (name, age) VALUES ('张三', 25)
  UPDATE my_index SET age=26 WHERE name='张三'
  DELETE FROM my_index WHERE name='张三'
  ALTER INDEX my_index SETTINGS (number_of_replicas=2)
  ALTER TABLE my_index RENAME TO new_index

支持反斜杠 (\) 续行:
  SELECT * FROM my_index \
  WHERE id = 1 \
  LIMIT 10`)
}

func (r *REPL) setFormat(format string) {
	f, err := replFmt.NewFormatter(format)
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

func (r *REPL) handlePing() {
	res, err := r.client.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: ping ES 失败: %v\n", err)
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		fmt.Fprintf(os.Stderr, "错误: ES ping 失败: %s\n", res.String())
		return
	}

	fmt.Printf("%s: pong\n", strings.Join(r.addresses, " "))
}
