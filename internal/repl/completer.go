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
	{Text: "CREATE TABLE", Description: "创建索引"},
	{Text: "DROP TABLE", Description: "删除索引"},
	{Text: "ALTER INDEX", Description: "修改索引设置"},
	{Text: "ALTER TABLE", Description: "修改索引"},
	{Text: "RENAME TO", Description: "重命名索引"},
	{Text: "SETTINGS", Description: "索引设置"},
}

var builtinCommands = []prompt.Suggest{
	{Text: ".help", Description: "显示帮助"},
	{Text: ".ping", Description: "测试 ES 连接"},
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
		strings.Contains(upperText, ".SCHEMA") ||
		strings.Contains(upperText, "CREATE TABLE") || strings.Contains(upperText, "DROP TABLE") ||
		strings.Contains(upperText, "ALTER INDEX") || strings.Contains(upperText, "ALTER TABLE") {
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
