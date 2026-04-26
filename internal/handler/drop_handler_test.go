package handler

import "testing"

func TestParseDropTable(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		wantTable string
		wantErr   bool
	}{
		{
			name:      "basic drop",
			sql:       "DROP TABLE users",
			wantTable: "users",
		},
		{
			name:      "case insensitive",
			sql:       "drop table users",
			wantTable: "users",
		},
		{
			name:      "with semicolon",
			sql:       "DROP TABLE users;",
			wantTable: "users",
		},
		{
			name:      "with leading spaces",
			sql:       "  DROP TABLE users  ",
			wantTable: "users",
		},
		{
			name:    "invalid sql",
			sql:     "NOT A VALID SQL",
			wantErr: true,
		},
		{
			name:    "missing table name",
			sql:     "DROP TABLE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, err := parseDropTable(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDropTable(%q) expected error, got nil", tt.sql)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseDropTable(%q) unexpected error: %v", tt.sql, err)
			}

			if tableName != tt.wantTable {
				t.Errorf("table name = %q, want %q", tableName, tt.wantTable)
			}
		})
	}
}
