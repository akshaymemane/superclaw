package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Hooks is a set of optional callbacks invoked at key points in the agent loop.
// Nil function fields are silently skipped — set only the events you care about.
type Hooks struct {
	// OnStep is called at the start of each iteration before the LLM call.
	OnStep func(step int)
	// OnToolCall is called just before a tool is executed.
	OnToolCall func(step int, name string, input json.RawMessage)
	// OnToolResult is called immediately after a tool returns.
	OnToolResult func(step int, name string, result string, isError bool)
}

// ProgressHooks returns Hooks that write a human-readable progress log to w.
// It is suitable for wiring to os.Stderr in a CLI.
func ProgressHooks(w io.Writer) Hooks {
	return Hooks{
		OnToolCall: func(step int, name string, input json.RawMessage) {
			preview := strings.TrimSpace(string(input))
			if len([]rune(preview)) > 120 {
				preview = string([]rune(preview)[:120]) + "…"
			}
			fmt.Fprintf(w, "  [%d] → %-14s %s\n", step, name, preview)
		},
		OnToolResult: func(step int, name string, result string, isError bool) {
			mark := "✓"
			if isError {
				mark = "✗"
			}
			preview := strings.TrimSpace(result)
			if len([]rune(preview)) > 120 {
				preview = string([]rune(preview)[:120]) + "…"
			}
			fmt.Fprintf(w, "  [%d] %s %s\n", step, mark, preview)
		},
	}
}
