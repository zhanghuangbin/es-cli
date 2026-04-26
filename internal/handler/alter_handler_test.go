package handler

import "testing"

func TestAlterSettingsRegex(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		wantIndex    string
		wantSettings string
		wantMatch    bool
	}{
		{
			name:         "basic settings",
			sql:          "ALTER INDEX users SETTINGS (number_of_replicas=2)",
			wantIndex:    "users",
			wantSettings: "number_of_replicas=2",
			wantMatch:    true,
		},
		{
			name:         "multiple settings",
			sql:          "ALTER INDEX users SETTINGS (number_of_replicas=2, refresh_interval=30s)",
			wantIndex:    "users",
			wantSettings: "number_of_replicas=2, refresh_interval=30s",
			wantMatch:    true,
		},
		{
			name:         "case insensitive",
			sql:          "alter index users settings (number_of_replicas=2)",
			wantIndex:    "users",
			wantSettings: "number_of_replicas=2",
			wantMatch:    true,
		},
		{
			name:         "with semicolon",
			sql:          "ALTER INDEX users SETTINGS (number_of_replicas=2);",
			wantIndex:    "users",
			wantSettings: "number_of_replicas=2",
			wantMatch:    true,
		},
		{
			name:      "no match",
			sql:       "NOT A VALID SQL",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := reAlterSettings.FindStringSubmatch(tt.sql)

			if tt.wantMatch {
				if matches == nil {
					t.Fatalf("expected match for %q, got nil", tt.sql)
				}
				if matches[1] != tt.wantIndex {
					t.Errorf("index = %q, want %q", matches[1], tt.wantIndex)
				}
				if matches[2] != tt.wantSettings {
					t.Errorf("settings = %q, want %q", matches[2], tt.wantSettings)
				}
			} else {
				if matches != nil {
					t.Errorf("expected no match for %q, got %v", tt.sql, matches)
				}
			}
		})
	}
}

func TestAlterRenameRegex(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		wantOld   string
		wantNew   string
		wantMatch bool
	}{
		{
			name:      "basic rename",
			sql:       "ALTER TABLE users RENAME TO users_v2",
			wantOld:   "users",
			wantNew:   "users_v2",
			wantMatch: true,
		},
		{
			name:      "case insensitive",
			sql:       "alter table users rename to users_v2",
			wantOld:   "users",
			wantNew:   "users_v2",
			wantMatch: true,
		},
		{
			name:      "with semicolon",
			sql:       "ALTER TABLE users RENAME TO users_v2;",
			wantOld:   "users",
			wantNew:   "users_v2",
			wantMatch: true,
		},
		{
			name:      "no match",
			sql:       "ALTER INDEX users SETTINGS (x=1)",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := reAlterRename.FindStringSubmatch(tt.sql)

			if tt.wantMatch {
				if matches == nil {
					t.Fatalf("expected match for %q, got nil", tt.sql)
				}
				if matches[1] != tt.wantOld {
					t.Errorf("old name = %q, want %q", matches[1], tt.wantOld)
				}
				if matches[2] != tt.wantNew {
					t.Errorf("new name = %q, want %q", matches[2], tt.wantNew)
				}
			} else {
				if matches != nil {
					t.Errorf("expected no match for %q, got %v", tt.sql, matches)
				}
			}
		})
	}
}
