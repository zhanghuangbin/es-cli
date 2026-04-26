package handler

import "testing"

func TestParseInsertInto(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		wantTable string
		wantCols  []string
		wantDoc   map[string]any
		wantErr   bool
	}{
		{
			name:      "basic insert",
			sql:       "INSERT INTO users (name, age) VALUES ('张三', 25)",
			wantTable: "users",
			wantCols:  []string{"name", "age"},
			wantDoc:   map[string]any{"name": "张三", "age": int64(25)},
		},
		{
			name:      "insert with boolean and null",
			sql:       "INSERT INTO users (name, active, score) VALUES ('test', true, null)",
			wantTable: "users",
			wantCols:  []string{"name", "active", "score"},
			wantDoc:   map[string]any{"name": "test", "active": true, "score": nil},
		},
		{
			name:      "insert with float",
			sql:       "INSERT INTO products (name, price) VALUES ('book', 3.14)",
			wantTable: "products",
			wantCols:  []string{"name", "price"},
			wantDoc:   map[string]any{"name": "book", "price": 3.14},
		},
		{
			name:      "case insensitive",
			sql:       "insert into users (name) values ('test')",
			wantTable: "users",
			wantCols:  []string{"name"},
			wantDoc:   map[string]any{"name": "test"},
		},
		{
			name:      "with semicolon",
			sql:       "INSERT INTO users (name) VALUES ('test');",
			wantTable: "users",
			wantCols:  []string{"name"},
			wantDoc:   map[string]any{"name": "test"},
		},
		{
			name:    "invalid sql",
			sql:     "NOT A VALID SQL",
			wantErr: true,
		},
		{
			name:    "column value mismatch",
			sql:     "INSERT INTO users (name, age) VALUES ('test')",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, cols, doc, err := parseInsertInto(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseInsertInto(%q) expected error, got nil", tt.sql)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseInsertInto(%q) unexpected error: %v", tt.sql, err)
			}

			if tableName != tt.wantTable {
				t.Errorf("table name = %q, want %q", tableName, tt.wantTable)
			}

			if len(cols) != len(tt.wantCols) {
				t.Fatalf("cols count = %d, want %d", len(cols), len(tt.wantCols))
			}
			for i, c := range tt.wantCols {
				if cols[i] != c {
					t.Errorf("cols[%d] = %q, want %q", i, cols[i], c)
				}
			}

			for k, wantV := range tt.wantDoc {
				gotV, ok := doc[k]
				if !ok {
					t.Errorf("missing key %q in doc", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("doc[%q] = %v (%T), want %v (%T)", k, gotV, gotV, wantV, wantV)
				}
			}
		})
	}
}

func TestParseInsertColumns(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"single column", "name", []string{"name"}, false},
		{"multiple columns", "name, age, active", []string{"name", "age", "active"}, false},
		{"with spaces", " name , age ", []string{"name", "age"}, false},
		{"empty", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInsertColumns(tt.input)
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
				t.Fatalf("got %d columns, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("columns[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSplitValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple", "'a', 'b'", []string{"'a'", "'b'"}},
		{"with comma in string", "'a,b', 'c'", []string{"'a,b'", "'c'"}},
		{"mixed types", "'hello', 42, true", []string{"'hello'", "42", "true"}},
		{"escaped quote", "'it''s', 'ok'", []string{"'it''s'", "'ok'"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitValues(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d values %v, want %d values %v", len(got), got, len(tt.want), tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("values[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    any
		wantErr bool
	}{
		{"string", "'hello'", "hello", false},
		{"escaped quote", "'it''s'", "it's", false},
		{"integer", "42", int64(42), false},
		{"negative integer", "-10", int64(-10), false},
		{"float", "3.14", 3.14, false},
		{"true", "true", true, false},
		{"TRUE", "TRUE", true, false},
		{"false", "false", false, false},
		{"null", "null", nil, false},
		{"NULL", "NULL", nil, false},
		{"empty", "", nil, true},
		{"invalid", "abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseValue(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseValue(%q) = %v (%T), want %v (%T)", tt.input, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestParseInsertValues(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []any
		wantErr bool
	}{
		{
			name:  "mixed types",
			input: "'hello', 42, true, null",
			want:  []any{"hello", int64(42), true, nil},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInsertValues(tt.input)
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
				t.Fatalf("got %d values, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("values[%d] = %v (%T), want %v (%T)", i, got[i], got[i], tt.want[i], tt.want[i])
				}
			}
		})
	}
}
