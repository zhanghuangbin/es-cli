package es

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigFields(t *testing.T) {
	cfg := Config{
		Addresses: []string{"http://localhost:9200", "http://localhost:9201"},
		Username:  "elastic",
		Password:  "changeme",
		CACert:    "/path/to/ca.crt",
	}

	if len(cfg.Addresses) != 2 {
		t.Errorf("expected 2 addresses, got %d", len(cfg.Addresses))
	}
	if cfg.Username != "elastic" {
		t.Errorf("username = %q, want 'elastic'", cfg.Username)
	}
	if cfg.Password != "changeme" {
		t.Errorf("password = %q, want 'changeme'", cfg.Password)
	}
	if cfg.CACert != "/path/to/ca.crt" {
		t.Errorf("caCert = %q, want '/path/to/ca.crt'", cfg.CACert)
	}
}

func TestNewClientBasic(t *testing.T) {
	cfg := Config{
		Addresses: []string{"http://localhost:9200"},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewClientWithAuth(t *testing.T) {
	cfg := Config{
		Addresses: []string{"http://localhost:9200"},
		Username:  "elastic",
		Password:  "changeme",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewClientWithInvalidCACert(t *testing.T) {
	cfg := Config{
		Addresses: []string{"http://localhost:9200"},
		CACert:    "/non/existent/path/ca.crt",
	}

	_, err := NewClient(cfg)
	if err == nil {
		t.Error("expected error for non-existent CA cert")
	}
}

func TestNewClientWithInvalidCACertContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid_ca.crt")
	if err := os.WriteFile(tmpFile, []byte("not a valid cert"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cfg := Config{
		Addresses: []string{"http://localhost:9200"},
		CACert:    tmpFile,
	}

	_, err := NewClient(cfg)
	if err == nil {
		t.Error("expected error for invalid CA cert content")
	}
}
