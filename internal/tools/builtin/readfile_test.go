package builtin_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/akshaymemane/superclaw/internal/tools/builtin"
)

func TestReadFileTool_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}

	tool := builtin.NewReadFileTool(dir)
	input, _ := json.Marshal(map[string]string{"path": path})
	out, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", out)
	}
}

func TestReadFileTool_OutsideBaseDir(t *testing.T) {
	dir := t.TempDir()
	tool := builtin.NewReadFileTool(dir)

	// Attempt to read a file outside the base dir.
	input, _ := json.Marshal(map[string]string{"path": "/etc/passwd"})
	_, err := tool.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for path outside base dir")
	}
}

func TestWriteFileTool_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "out.txt")

	wt := builtin.NewWriteFileTool(dir)
	wInput, _ := json.Marshal(map[string]string{"path": path, "content": "superclaw"})
	if _, err := wt.Execute(context.Background(), wInput); err != nil {
		t.Fatalf("write error: %v", err)
	}

	rt := builtin.NewReadFileTool(dir)
	rInput, _ := json.Marshal(map[string]string{"path": path})
	out, err := rt.Execute(context.Background(), rInput)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if out != "superclaw" {
		t.Errorf("expected %q, got %q", "superclaw", out)
	}
}
