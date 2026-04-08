package session_test

import (
	"errors"
	"testing"

	"github.com/akshaymemane/superclaw/internal/session"
)

func TestStoreAppendAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)

	// Load from non-existent file returns empty slice, not error.
	records, err := store.Load()
	if err != nil {
		t.Fatalf("Load on empty store: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}

	r1 := session.NewRecord("task one", "result one", 3, "success", nil)
	if err := store.Append(r1); err != nil {
		t.Fatalf("Append r1: %v", err)
	}

	r2 := session.NewRecord("task two", "", 1, "error", errors.New("timeout"))
	if err := store.Append(r2); err != nil {
		t.Fatalf("Append r2: %v", err)
	}

	records, err = store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	if records[0].Task != "task one" {
		t.Errorf("records[0].Task = %q, want %q", records[0].Task, "task one")
	}
	if records[0].Steps != 3 {
		t.Errorf("records[0].Steps = %d, want 3", records[0].Steps)
	}
	if records[0].Status != "success" {
		t.Errorf("records[0].Status = %q, want success", records[0].Status)
	}
	if records[0].Error != "" {
		t.Errorf("records[0].Error = %q, want empty", records[0].Error)
	}

	if records[1].Task != "task two" {
		t.Errorf("records[1].Task = %q, want %q", records[1].Task, "task two")
	}
	if records[1].Status != "error" {
		t.Errorf("records[1].Status = %q, want error", records[1].Status)
	}
	if records[1].Error != "timeout" {
		t.Errorf("records[1].Error = %q, want timeout", records[1].Error)
	}
}

func TestNewRecord(t *testing.T) {
	r := session.NewRecord("my task", "my result", 5, "success", nil)

	if r.ID == "" {
		t.Error("ID must not be empty")
	}
	if r.Timestamp.IsZero() {
		t.Error("Timestamp must not be zero")
	}
	if r.Task != "my task" {
		t.Errorf("Task = %q, want %q", r.Task, "my task")
	}
	if r.Result != "my result" {
		t.Errorf("Result = %q, want %q", r.Result, "my result")
	}
	if r.Steps != 5 {
		t.Errorf("Steps = %d, want 5", r.Steps)
	}
	if r.Status != "success" {
		t.Errorf("Status = %q, want success", r.Status)
	}
}
