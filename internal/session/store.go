// Package session persists agent run records to a JSONL transcript store.
// Records are appended to .superclaw/runs.jsonl inside the work directory.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	storeSubdir = ".superclaw"
	storeFile   = "runs.jsonl"
)

// RunRecord captures the outcome of a single agent run for persistence.
type RunRecord struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Task      string    `json:"task"`
	Result    string    `json:"result"`
	Steps     int       `json:"steps"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
}

// Store appends run records to a JSONL file.
type Store struct {
	path string
}

// NewStore returns a Store rooted at workDir.
// The store file is created on first Append if it does not exist.
func NewStore(workDir string) *Store {
	return &Store{
		path: filepath.Join(workDir, storeSubdir, storeFile),
	}
}

// Append serialises r and appends it as a single JSON line.
func (s *Store) Append(r RunRecord) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("session store: %w", err)
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("session store: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(r)
}

// Load reads all records from the store. Returns nil, nil when the file does
// not exist yet.
func (s *Store) Load() ([]RunRecord, error) {
	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("session store: %w", err)
	}
	defer f.Close()

	var records []RunRecord
	dec := json.NewDecoder(f)
	for dec.More() {
		var r RunRecord
		if err := dec.Decode(&r); err != nil {
			return records, fmt.Errorf("session store: corrupt record: %w", err)
		}
		records = append(records, r)
	}
	return records, nil
}

// NewRecord creates a RunRecord stamped with the current time.
func NewRecord(task, result string, steps int, status string, runErr error) RunRecord {
	r := RunRecord{
		ID:        fmt.Sprintf("%s", time.Now().Format("20060102-150405.000")),
		Timestamp: time.Now().UTC(),
		Task:      task,
		Result:    result,
		Steps:     steps,
		Status:    status,
	}
	if runErr != nil {
		r.Error = runErr.Error()
	}
	return r
}
