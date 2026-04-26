package repl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryAdd(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
	}

	h.Add("SELECT * FROM users")
	h.Add("SHOW TABLES")

	entries := h.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0] != "SELECT * FROM users" {
		t.Errorf("entries[0] = %q, want 'SELECT * FROM users'", entries[0])
	}
	if entries[1] != "SHOW TABLES" {
		t.Errorf("entries[1] = %q, want 'SHOW TABLES'", entries[1])
	}
}

func TestHistoryAddEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
	}

	h.Add("")
	h.Add("  ")

	if len(h.Entries()) != 0 {
		t.Errorf("expected 0 entries, got %d", len(h.Entries()))
	}
}

func TestHistoryAddDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
	}

	h.Add("SELECT 1")
	h.Add("SELECT 1")

	if len(h.Entries()) != 1 {
		t.Errorf("expected 1 entry (deduplicated), got %d", len(h.Entries()))
	}
}

func TestHistoryAddNonConsecutiveDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	h := &History{
		filePath: filepath.Join(tmpDir, "history"),
	}

	h.Add("SELECT 1")
	h.Add("SELECT 2")
	h.Add("SELECT 1") // 非连续重复应该保留

	if len(h.Entries()) != 3 {
		t.Errorf("expected 3 entries, got %d", len(h.Entries()))
	}
}

func TestHistoryPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	// 写入
	h1 := &History{
		filePath: historyFile,
	}
	h1.Add("SELECT 1")
	h1.Add("SELECT 2")

	// 读取
	h2 := &History{
		filePath: historyFile,
	}
	h2.load()

	entries := h2.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after reload, got %d", len(entries))
	}
	if entries[0] != "SELECT 1" {
		t.Errorf("entries[0] = %q, want 'SELECT 1'", entries[0])
	}
}

func TestHistoryLoadNonExistent(t *testing.T) {
	h := &History{
		filePath: filepath.Join(os.TempDir(), "non_existent_history_file_test"),
	}
	h.load()

	if len(h.Entries()) != 0 {
		t.Errorf("expected 0 entries for non-existent file, got %d", len(h.Entries()))
	}
}
