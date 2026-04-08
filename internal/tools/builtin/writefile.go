package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteFileTool creates or overwrites a file, constrained to a base directory.
type WriteFileTool struct {
	BaseDir string
}

func NewWriteFileTool(baseDir string) *WriteFileTool {
	return &WriteFileTool{BaseDir: baseDir}
}

func (w *WriteFileTool) Name() string { return "write_file" }
func (w *WriteFileTool) Description() string {
	return "Write content to a local file, creating parent directories as needed. Overwrites existing files."
}
func (w *WriteFileTool) Properties() map[string]any {
	return map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Destination file path.",
		},
		"content": map[string]any{
			"type":        "string",
			"description": "Content to write.",
		},
	}
}

type writeFileInput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func (w *WriteFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var in writeFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	clean, err := w.safePath(in.Path)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
		return "", fmt.Errorf("cannot create directories: %w", err)
	}
	if err := os.WriteFile(clean, []byte(in.Content), 0o644); err != nil {
		return "", fmt.Errorf("cannot write file: %w", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(in.Content), in.Path), nil
}

func (w *WriteFileTool) safePath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path is required")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if w.BaseDir != "" {
		base, err := filepath.Abs(w.BaseDir)
		if err != nil {
			return "", fmt.Errorf("invalid base dir: %w", err)
		}
		if abs != base && !strings.HasPrefix(abs, base+string(filepath.Separator)) {
			return "", fmt.Errorf("path %q is outside the allowed directory", p)
		}
	}
	return abs, nil
}
