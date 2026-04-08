package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const readMaxBytes = 64 * 1024 // 64 KB

// ReadFileTool reads files from disk, constrained to a base directory.
type ReadFileTool struct {
	BaseDir string
}

func NewReadFileTool(baseDir string) *ReadFileTool {
	return &ReadFileTool{BaseDir: baseDir}
}

func (r *ReadFileTool) Name() string { return "read_file" }
func (r *ReadFileTool) Description() string {
	return "Read the contents of a local file. Returns up to 64 KB."
}
func (r *ReadFileTool) Properties() map[string]any {
	return map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to the file to read.",
		},
	}
}

type readFileInput struct {
	Path string `json:"path"`
}

func (r *ReadFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var in readFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	clean, err := r.safePath(in.Path)
	if err != nil {
		return "", err
	}

	f, err := os.Open(clean)
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %w", in.Path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, readMaxBytes+1))
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", in.Path, err)
	}

	content := string(data)
	if len(data) > readMaxBytes {
		content = string(data[:readMaxBytes]) + "\n[truncated]"
	}
	return content, nil
}

func (r *ReadFileTool) safePath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path is required")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if r.BaseDir != "" {
		base, err := filepath.Abs(r.BaseDir)
		if err != nil {
			return "", fmt.Errorf("invalid base dir: %w", err)
		}
		if abs != base && !strings.HasPrefix(abs, base+string(filepath.Separator)) {
			return "", fmt.Errorf("path %q is outside the allowed directory", p)
		}
	}
	return abs, nil
}
