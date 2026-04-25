package executor

import "testing"

func TestParseCreateTable(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		wantTable     string
		wantColCount  int
		wantSettings  bool
		wantErr       bool
	}{
		{
			name:         "basic two columns",
			sql:          "CREATE TABLE users (name TEXT, age INTEGER)",
			wantTable:    "users",
			wantColCount: 2,
			wantSettings: false,
		},
		{
			name:         "with settings",
			sql:          "CREATE TABLE users (name TEXT) SETTINGS (number_of_shards=3)",
			wantTable:    "users",
			wantColCount: 1,
			wantSettings: true,
		},
		{
			name:         "case insensitive",
			sql:          "create table Users (Name text)",
			wantTable:    "Users",
			wantColCount: 1,
			wantSettings: false,
		},
		{
			name:         "with semicolon",
			sql:          "CREATE TABLE users (name TEXT);",
			wantTable:    "users",
			wantColCount: 1,
			wantSettings: false,
		},
		{
			name:    "invalid sql",
			sql:     "NOT A VALID SQL",
			wantErr: true,
		},
		{
			name:    "missing columns",
			sql:     "CREATE TABLE users ()",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, columns, settings, err := parseCreateTable(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseCreateTable(%q) expected error, got nil", tt.sql)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseCreateTable(%q) unexpected error: %v", tt.sql, err)
			}

			if tableName != tt.wantTable {
				t.Errorf("table name = %q, want %q", tableName, tt.wantTable)
			}

			if len(columns) != tt.wantColCount {
				t.Errorf("column count = %d, want %d", len(columns), tt.wantColCount)
			}

			if tt.wantSettings && settings == nil {
				t.Error("expected settings to be non-nil")
			}
			if !tt.wantSettings && settings != nil {
				t.Errorf("expected settings to be nil, got %v", settings)
			}
		})
	}
}

func TestParseColumns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCols []columnDef
		wantErr  bool
	}{
		{
			name:  "multiple columns",
			input: "name TEXT, age INTEGER, active BOOLEAN",
			wantCols: []columnDef{
				{Name: "name", Type: "text"},
				{Name: "age", Type: "integer"},
				{Name: "active", Type: "boolean"},
			},
		},
		{
			name:  "all int types",
			input: "a INT, b INTEGER, c BIGINT, d LONG",
			wantCols: []columnDef{
				{Name: "a", Type: "integer"},
				{Name: "b", Type: "integer"},
				{Name: "c", Type: "long"},
				{Name: "d", Type: "long"},
			},
		},
		{
			name:  "float and double",
			input: "a FLOAT, b DOUBLE",
			wantCols: []columnDef{
				{Name: "a", Type: "float"},
				{Name: "b", Type: "double"},
			},
		},
		{
			name:  "boolean variants",
			input: "a BOOLEAN, b BOOL",
			wantCols: []columnDef{
				{Name: "a", Type: "boolean"},
				{Name: "b", Type: "boolean"},
			},
		},
		{
			name:  "string types",
			input: "a TEXT, b KEYWORD, c VARCHAR",
			wantCols: []columnDef{
				{Name: "a", Type: "text"},
				{Name: "b", Type: "keyword"},
				{Name: "c", Type: "keyword"},
			},
		},
		{
			name:  "date types",
			input: "a DATE, b DATETIME, c TIMESTAMP",
			wantCols: []columnDef{
				{Name: "a", Type: "date"},
				{Name: "b", Type: "date"},
				{Name: "c", Type: "date"},
			},
		},
		{
			name:  "complex types",
			input: "a OBJECT, b JSON, c NESTED",
			wantCols: []columnDef{
				{Name: "a", Type: "object"},
				{Name: "b", Type: "object"},
				{Name: "c", Type: "nested"},
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing type",
			input:   "name",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   "name BLOB",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols, err := parseColumns(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseColumns(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseColumns(%q) unexpected error: %v", tt.input, err)
			}

			if len(cols) != len(tt.wantCols) {
				t.Fatalf("got %d columns, want %d", len(cols), len(tt.wantCols))
			}

			for i, want := range tt.wantCols {
				if cols[i].Name != want.Name {
					t.Errorf("column[%d].Name = %q, want %q", i, cols[i].Name, want.Name)
				}
				if cols[i].Type != want.Type {
					t.Errorf("column[%d].Type = %q, want %q", i, cols[i].Type, want.Type)
				}
			}
		})
	}
}

func TestParseSettings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantMap  map[string]string
		wantErr  bool
	}{
		{
			name:  "single setting",
			input: "number_of_shards=3",
			wantMap: map[string]string{
				"number_of_shards": "3",
			},
		},
		{
			name:  "multiple settings",
			input: "number_of_shards=3, number_of_replicas=1",
			wantMap: map[string]string{
				"number_of_shards":    "3",
				"number_of_replicas": "1",
			},
		},
		{
			name:    "missing equals",
			input:   "number_of_shards",
			wantErr: true,
		},
		{
			name:    "empty key",
			input:   "=3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := parseSettings(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSettings(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseSettings(%q) unexpected error: %v", tt.input, err)
			}

			if len(settings) != len(tt.wantMap) {
				t.Fatalf("got %d settings, want %d", len(settings), len(tt.wantMap))
			}

			for k, wantV := range tt.wantMap {
				gotV, ok := settings[k]
				if !ok {
					t.Errorf("missing key %q", k)
					continue
				}
				if gotV != wantV {
					t.Errorf("settings[%q] = %q, want %q", k, gotV, wantV)
				}
			}
		})
	}
}
