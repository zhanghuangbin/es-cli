package types

// Meta 描述 SQL 执行的元信息。
type Meta struct {
	Status  int
	Type    SQLType // SQL 语句类型
	Message string
	Stat    map[string]any // ES 处理结果的统计信息，如查询耗时、版本号、影响行数等
}

// Result 表示一条 SQL 语句的执行结果。
type Result struct {
	Meta    Meta
	Columns []string
	Rows    [][]any
}
