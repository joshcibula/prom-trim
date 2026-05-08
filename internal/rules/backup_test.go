package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupFile_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "rules.yaml")
	content := []byte("groups: []\n")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	dst, err := BackupFile(src, BackupOptions{Suffix: ".bak"})
	if err != nil {
		t.Fatalf("BackupFile: %v", err)
	}

	if !strings.HasSuffix(dst, ".bak") {
		t.Errorf("expected .bak suffix, got %s", dst)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("backup content mismatch: got %q want %q", got, content)
	}
}

func TestBackupFile_DefaultSuffix(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(src, []byte("groups: []\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	dst, err := BackupFile(src, BackupOptions{})
	if err != nil {
		t.Fatalf("BackupFile: %v", err)
	}

	if !strings.HasSuffix(dst, ".bak") {
		t.Errorf("expected timestamp .bak suffix, got %s", dst)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Errorf("backup file does not exist: %s", dst)
	}
}

func TestBackupFile_SourceNotFound(t *testing.T) {
	_, err := BackupFile("/nonexistent/path/rules.yaml", BackupOptions{Suffix: ".bak"})
	if err == nil {
		t.Fatal("expected error for missing source file, got nil")
	}
}

func TestBackupFile_OriginalUnchanged(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "rules.yaml")
	content := []byte("groups: []\n")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := BackupFile(src, BackupOptions{Suffix: ".bak"}); err != nil {
		t.Fatalf("BackupFile: %v", err)
	}

	got, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("original modified: got %q want %q", got, content)
	}
}
