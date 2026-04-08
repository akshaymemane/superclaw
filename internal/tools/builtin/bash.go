package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const bashDefaultTimeout = 30 * time.Second

// BashTool executes shell commands with a timeout and a blocklist of
// dangerous patterns. Use the blocklist to prevent the most obviously
// destructive operations; it is not a full sandbox.
type BashTool struct {
	Timeout         time.Duration
	BlockedPatterns []string
}

func NewBashTool() *BashTool {
	return &BashTool{
		Timeout: bashDefaultTimeout,
		BlockedPatterns: []string{
			"rm -rf /",
			"dd if=",
			"mkfs",
			":(){:|:&};:",
			"chmod -R 777 /",
			"> /dev/sda",
		},
	}
}

func (b *BashTool) Name() string { return "run_bash" }
func (b *BashTool) Description() string {
	return "Run a shell command and return its combined stdout/stderr output. Times out after 30 seconds."
}
func (b *BashTool) Properties() map[string]any {
	return map[string]any{
		"command": map[string]any{
			"type":        "string",
			"description": "The shell command to execute.",
		},
	}
}

type bashInput struct {
	Command string `json:"command"`
}

func (b *BashTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in bashInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if in.Command == "" {
		return "", fmt.Errorf("command is required")
	}

	for _, pattern := range b.BlockedPatterns {
		if strings.Contains(in.Command, pattern) {
			return "", fmt.Errorf("command contains blocked pattern %q", pattern)
		}
	}

	timeout := b.Timeout
	if timeout == 0 {
		timeout = bashDefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", in.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	out := stdout.String()
	if stderr.Len() > 0 {
		out += "\nSTDERR:\n" + stderr.String()
	}
	if runErr != nil {
		return out, fmt.Errorf("command failed: %w", runErr)
	}
	return out, nil
}
