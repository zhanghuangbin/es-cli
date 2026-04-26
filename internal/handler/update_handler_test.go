package handler

import (
	"encoding/json"
	"testing"
)

func TestParseUpdate(t *testing.T) {
	tests := []struct {
		name           string
		sql            string
		wantTable      string
		wantSetCount   int
		wantWhereCount int
		wantErr        bool
	}{
		{
			name:           "basic update",
			sql:            "UPDATE users SET name='test' WHERE id=1",
			wantTable:      "users",
			wantSetCount:   1,
			wantWhereCount: 1,
		},
		{
			name:           "multiple set and where",
			sql:            "UPDATE users SET name='test', age=30 WHERE id=1 AND active=true",
			wantTable:      "users",
			wantSetCount:   2,
			wantWhereCount: 2,
		},
		{
			name:           "case insensitive",
			sql:            "update users set name='test' where id=1",
			wantTable:      "users",
			wantSetCount:   1,
			wantWhereCount: 1,
		},
		{
			name:           "with semicolon",
			sql:            "UPDATE users SET name='test' WHERE id=1;",
			wantTable:      "users",
			wantSetCount:   1,
			wantWhereCount: 1,
		},
		{
			name:    "invalid sql",
			sql:     "NOT A VALID SQL",
			wantErr: true,
		},
		{
			name:    "missing where",
			sql:     "UPDATE users SET name='test'",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, setFields, whereConditions, err := parseUpdate(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseUpdate(%q) expected error, got nil", tt.sql)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseUpdate(%q) unexpected error: %v", tt.sql, err)
			}

			if tableName != tt.wantTable {
				t.Errorf("table name = %q, want %q", tableName, tt.wantTable)
			}

			if len(setFields) != tt.wantSetCount {
				t.Errorf("set fields count = %d, want %d", len(setFields), tt.wantSetCount)
			}

			if len(whereConditions) != tt.wantWhereCount {
				t.Errorf("where conditions count = %d, want %d", len(whereConditions), tt.wantWhereCount)
			}
		})
	}
}

func TestParseSetClause(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]any
		wantErr bool
	}{
		{
			name:  "single field string",
			input: "name='hello'",
			want:  map[string]any{"name": "hello"},
		},
		{
			name:  "single field integer",
			input: "age=42",
			want:  map[string]any{"age": int64(42)},
		},
		{
			name:  "multiple fields",
			input: "name='test', age=30",
			want:  map[string]any{"name": "test", "age": int64(30)},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing equals",
			input:   "name",
			wantErr: true,
		},
		{
			name:    "empty key",
			input:   "='hello'",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSetClause(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d fields, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("fields[%q] = %v (%T), want %v (%T)", k, gotV, gotV, wantV, wantV)
				}
			}
		})
	}
}

func TestParseWhereClause(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]any
		wantErr bool
	}{
		{
			name:  "single condition",
			input: "id=1",
			want:  map[string]any{"id": int64(1)},
		},
		{
			name:  "multiple conditions",
			input: "name='test' AND age=30",
			want:  map[string]any{"name": "test", "age": int64(30)},
		},
		{
			name:  "case insensitive AND",
			input: "name='test' and age=30",
			want:  map[string]any{"name": "test", "age": int64(30)},
		},
		{
			name:  "boolean value",
			input: "active=true",
			want:  map[string]any{"active": true},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing equals",
			input:   "id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWhereClause(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d conditions, want %d", len(got), len(tt.want))
			}
			for k, wantV := range tt.want {
				gotV, ok := got[k]
				if !ok {
					t.Errorf("missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("conditions[%q] = %v (%T), want %v (%T)", k, gotV, gotV, wantV, wantV)
				}
			}
		})
	}
}

func TestBuildUpdateByQueryBody(t *testing.T) {
	setFields := map[string]any{
		"name": "test",
	}
	whereConditions := map[string]any{
		"id": int64(1),
	}

	body, err := buildUpdateByQueryBody(setFields, whereConditions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// 验证 script 存在
	script, ok := parsed["script"].(map[string]any)
	if !ok {
		t.Fatal("missing script in body")
	}
	if _, ok := script["source"]; !ok {
		t.Error("missing script.source")
	}
	if _, ok := script["params"]; !ok {
		t.Error("missing script.params")
	}

	// 验证 query 存在
	query, ok := parsed["query"].(map[string]any)
	if !ok {
		t.Fatal("missing query in body")
	}
	boolQuery, ok := query["bool"].(map[string]any)
	if !ok {
		t.Fatal("missing query.bool")
	}
	if _, ok := boolQuery["must"]; !ok {
		t.Error("missing query.bool.must")
	}
}
