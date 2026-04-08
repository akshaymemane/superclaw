package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/akshaymemane/superclaw/internal/config"
)

func TestLoad_RoundTrip(t *testing.T) {
	f := &config.File{
		Model:        "claude-opus-4-6",
		MaxSteps:     10,
		MaxTokens:    8000,
		WorkDir:      "/tmp/work",
		Tools:        []string{"fetch_url", "read_file"},
		Skills:       []string{"summarize"},
		SystemPrompt: "Be brief.",
	}

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	path := filepath.Join(t.TempDir(), "superclaw.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Model != f.Model {
		t.Errorf("Model: want %q, got %q", f.Model, got.Model)
	}
	if got.MaxSteps != f.MaxSteps {
		t.Errorf("MaxSteps: want %d, got %d", f.MaxSteps, got.MaxSteps)
	}
	if len(got.Tools) != len(f.Tools) {
		t.Errorf("Tools len: want %d, got %d", len(f.Tools), len(got.Tools))
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := config.Load("/no/such/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadOrDefault_MissingReturnsEmpty(t *testing.T) {
	f, err := config.LoadOrDefault("/no/such/file.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil File")
	}
	if f.Model != "" {
		t.Errorf("expected empty model, got %q", f.Model)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
