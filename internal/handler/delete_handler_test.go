package handler

import (
	"encoding/json"
	"testing"
)

func TestParseDeleteFrom(t *testing.T) {
	tests := []struct {
		name           string
		sql            string
		wantTable      string
		wantWhereCount int
		wantErr        bool
	}{
		{
			name:           "basic delete",
			sql:            "DELETE FROM users WHERE id=1",
			wantTable:      "users",
			wantWhereCount: 1,
		},
		{
			name:           "multiple conditions",
			sql:            "DELETE FROM users WHERE name='test' AND age=30",
			wantTable:      "users",
			wantWhereCount: 2,
		},
		{
			name:           "case insensitive",
			sql:            "delete from users where id=1",
			wantTable:      "users",
			wantWhereCount: 1,
		},
		{
			name:           "with semicolon",
			sql:            "DELETE FROM users WHERE id=1;",
			wantTable:      "users",
			wantWhereCount: 1,
		},
		{
			name:    "invalid sql",
			sql:     "NOT A VALID SQL",
			wantErr: true,
		},
		{
			name:    "missing where",
			sql:     "DELETE FROM users",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, whereConditions, err := parseDeleteFrom(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDeleteFrom(%q) expected error, got nil", tt.sql)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseDeleteFrom(%q) unexpected error: %v", tt.sql, err)
			}

			if tableName != tt.wantTable {
				t.Errorf("table name = %q, want %q", tableName, tt.wantTable)
			}

			if len(whereConditions) != tt.wantWhereCount {
				t.Errorf("where conditions count = %d, want %d", len(whereConditions), tt.wantWhereCount)
			}
		})
	}
}

func TestBuildDeleteByQueryBody(t *testing.T) {
	whereConditions := map[string]any{
		"name": "test",
		"age":  int64(30),
	}

	body, err := buildDeleteByQueryBody(whereConditions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	query, ok := parsed["query"].(map[string]any)
	if !ok {
		t.Fatal("missing query in body")
	}

	boolQuery, ok := query["bool"].(map[string]any)
	if !ok {
		t.Fatal("missing query.bool")
	}

	must, ok := boolQuery["must"].([]any)
	if !ok {
		t.Fatal("missing query.bool.must")
	}

	if len(must) != 2 {
		t.Errorf("must clauses count = %d, want 2", len(must))
	}
}
