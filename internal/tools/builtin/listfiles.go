package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

const listMaxEntries = 500

// ListFilesTool lists files in a directory with optional glob filtering and
// depth limiting. It is the read-only complement to ReadFileTool.
type ListFilesTool struct {
	BaseDir string
}

func NewListFilesTool(baseDir string) *ListFilesTool {
	return &ListFilesTool{BaseDir: baseDir}
}

func (l *ListFilesTool) Name() string { return "list_files" }
func (l *ListFilesTool) Description() string {
	return "List files and directories under a path. Supports glob pattern filtering and depth limiting."
}
func (l *ListFilesTool) Properties() map[string]any {
	return map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Directory to list. Defaults to the base work directory when empty.",
		},
		"pattern": map[string]any{
			"type":        "string",
			"description": `Glob pattern to filter filenames, e.g. "*.go" or "*.md". Matches all files when empty.`,
		},
		"max_depth": map[string]any{
			"type":        "integer",
			"description": "Maximum recursion depth. 1 = immediate children only. 0 = unlimited.",
		},
	}
}

type listInput struct {
	Path     string `json:"path"`
	Pattern  string `json:"pattern"`
	MaxDepth int    `json:"max_depth"`
}

func (l *ListFilesTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var in listInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	root, err := l.safePath(in.Path)
	if err != nil {
		return "", err
	}

	var entries []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil
		}

		// Depth check.
		if in.MaxDepth > 0 {
			depth := strings.Count(rel, string(filepath.Separator)) + 1
			if depth > in.MaxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Pattern filter (applies to filename only).
		if in.Pattern != "" {
			matched, err := filepath.Match(in.Pattern, d.Name())
			if err != nil {
				return fmt.Errorf("invalid pattern %q: %w", in.Pattern, err)
			}
			if !matched && !d.IsDir() {
				return nil
			}
		}

		line := rel
		if d.IsDir() {
			line += "/"
		}
		entries = append(entries, line)

		if len(entries) >= listMaxEntries {
			return fmt.Errorf("listing truncated at %d entries", listMaxEntries)
		}
		return nil
	})

	// A truncation error is informational, not fatal.
	truncMsg := ""
	if err != nil && strings.HasPrefix(err.Error(), "listing truncated") {
		truncMsg = "\n" + err.Error()
		err = nil
	}
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "(empty)", nil
	}
	return strings.Join(entries, "\n") + truncMsg, nil
}

func (l *ListFilesTool) safePath(p string) (string, error) {
	if p == "" {
		if l.BaseDir != "" {
			return filepath.Abs(l.BaseDir)
		}
		return filepath.Abs(".")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if l.BaseDir != "" {
		base, err := filepath.Abs(l.BaseDir)
		if err != nil {
			return "", fmt.Errorf("invalid base dir: %w", err)
		}
		if abs != base && !strings.HasPrefix(abs, base+string(filepath.Separator)) {
			return "", fmt.Errorf("path %q is outside the allowed directory", p)
		}
	}
	return abs, nil
}
