package types

import "strings"

// SQLType 表示 SQL 语句的类型。
type SQLType int

const (
	SQLTypeSelect SQLType = iota
	SQLTypeInsert
	SQLTypeUpdate
	SQLTypeDelete
	SQLTypeCreate
	SQLTypeDrop
	SQLTypeAlter
)

var sqlTypeNames = map[SQLType]string{
	SQLTypeSelect: "SELECT",
	SQLTypeInsert: "INSERT",
	SQLTypeUpdate: "UPDATE",
	SQLTypeDelete: "DELETE",
	SQLTypeCreate: "CREATE",
	SQLTypeDrop:   "DROP",
	SQLTypeAlter:  "ALTER",
}

func (t SQLType) String() string {
	if name, ok := sqlTypeNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// DetectSQLType 根据 SQL 语句的前缀判断语句类型。
// 未知类型 fallback 到 SQLTypeSelect（走 _sql API）。
func DetectSQLType(sql string) SQLType {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return SQLTypeSelect
	}

	upper := strings.ToUpper(trimmed)

	switch {
	case strings.HasPrefix(upper, "SELECT"):
		return SQLTypeSelect
	case strings.HasPrefix(upper, "INSERT"):
		return SQLTypeInsert
	case strings.HasPrefix(upper, "UPDATE"):
		return SQLTypeUpdate
	case strings.HasPrefix(upper, "DELETE"):
		return SQLTypeDelete
	case strings.HasPrefix(upper, "CREATE"):
		return SQLTypeCreate
	case strings.HasPrefix(upper, "DROP"):
		return SQLTypeDrop
	case strings.HasPrefix(upper, "ALTER"):
		return SQLTypeAlter
	default:
		return SQLTypeSelect
	}
}
