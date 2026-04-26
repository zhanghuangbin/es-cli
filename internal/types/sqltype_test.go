package types

import "testing"

func TestDetectSQLType(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected SQLType
	}{
		// SELECT variants
		{"select basic", "SELECT * FROM users", SQLTypeSelect},
		{"select case insensitive", "select count(*) from orders", SQLTypeSelect},
		{"select leading whitespace", "  SELECT * FROM users", SQLTypeSelect},

		// INSERT
		{"insert", "INSERT INTO users (name) VALUES ('test')", SQLTypeInsert},

		// UPDATE
		{"update", "UPDATE users SET name='test' WHERE id=1", SQLTypeUpdate},

		// DELETE
		{"delete", "DELETE FROM users WHERE id=1", SQLTypeDelete},

		// CREATE
		{"create", "CREATE TABLE users (name TEXT)", SQLTypeCreate},

		// DROP
		{"drop", "DROP TABLE users", SQLTypeDrop},

		// ALTER
		{"alter", "ALTER INDEX users SETTINGS (number_of_replicas=2)", SQLTypeAlter},

		// Unknown / fallback to Select
		{"show tables fallback", "SHOW TABLES", SQLTypeSelect},
		{"describe fallback", "DESCRIBE users", SQLTypeSelect},
		{"empty string fallback", "", SQLTypeSelect},
		{"random text fallback", "something random", SQLTypeSelect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectSQLType(tt.sql)
			if got != tt.expected {
				t.Errorf("DetectSQLType(%q) = %v, want %v", tt.sql, got, tt.expected)
			}
		})
	}
}

func TestSQLTypeString(t *testing.T) {
	tests := []struct {
		sqlType  SQLType
		expected string
	}{
		{SQLTypeSelect, "SELECT"},
		{SQLTypeInsert, "INSERT"},
		{SQLTypeUpdate, "UPDATE"},
		{SQLTypeDelete, "DELETE"},
		{SQLTypeCreate, "CREATE"},
		{SQLTypeDrop, "DROP"},
		{SQLTypeAlter, "ALTER"},
		{SQLType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.sqlType.String()
			if got != tt.expected {
				t.Errorf("SQLType(%d).String() = %q, want %q", tt.sqlType, got, tt.expected)
			}
		})
	}
}
