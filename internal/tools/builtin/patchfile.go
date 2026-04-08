package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PatchFileTool applies a search-and-replace edit to a file.
// It requires old_text to appear exactly once, preventing accidental
// multi-site edits and ensuring the agent has read the file first.
type PatchFileTool struct {
	BaseDir string
}

func NewPatchFileTool(baseDir string) *PatchFileTool {
	return &PatchFileTool{BaseDir: baseDir}
}

func (p *PatchFileTool) Name() string { return "patch_file" }
func (p *PatchFileTool) Description() string {
	return "Replace an exact string in a file with new text. old_text must appear exactly once. Prefer this over write_file for targeted edits."
}
func (p *PatchFileTool) Properties() map[string]any {
	return map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to the file to patch.",
		},
		"old_text": map[string]any{
			"type":        "string",
			"description": "The exact text to find. Must match exactly once in the file.",
		},
		"new_text": map[string]any{
			"type":        "string",
			"description": "The replacement text.",
		},
	}
}

type patchInput struct {
	Path    string `json:"path"`
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

func (p *PatchFileTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var in patchInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if in.Path == "" {
		return "", fmt.Errorf("path is required")
	}
	if in.OldText == "" {
		return "", fmt.Errorf("old_text is required")
	}

	clean, err := p.safePath(in.Path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(clean)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", in.Path, err)
	}
	content := string(data)

	count := strings.Count(content, in.OldText)
	switch count {
	case 0:
		return "", fmt.Errorf("old_text not found in %s", in.Path)
	case 1:
		// exactly one match — safe to replace
	default:
		return "", fmt.Errorf("old_text matches %d times in %s; it must be unique", count, in.Path)
	}

	patched := strings.Replace(content, in.OldText, in.NewText, 1)
	if err := os.WriteFile(clean, []byte(patched), 0o644); err != nil {
		return "", fmt.Errorf("cannot write %s: %w", in.Path, err)
	}

	delta := len(patched) - len(content)
	sign := "+"
	if delta < 0 {
		sign = ""
	}
	return fmt.Sprintf("patched %s (%s%d bytes)", in.Path, sign, delta), nil
}

func (p *PatchFileTool) safePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if p.BaseDir != "" {
		base, err := filepath.Abs(p.BaseDir)
		if err != nil {
			return "", fmt.Errorf("invalid base dir: %w", err)
		}
		if abs != base && !strings.HasPrefix(abs, base+string(filepath.Separator)) {
			return "", fmt.Errorf("path %q is outside the allowed directory", path)
		}
	}
	return abs, nil
}
