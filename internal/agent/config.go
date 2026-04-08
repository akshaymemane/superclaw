package agent

import (
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

const (
	DefaultModel     = anthropic.ModelClaudeOpus4_6
	DefaultMaxSteps  = 20
	DefaultMaxTokens = int64(16000)
)

// Config controls the agent's model, resource limits, and base instructions.
type Config struct {
	// Model is the Anthropic model to use.
	Model anthropic.Model
	// MaxSteps is the maximum number of LLM calls before the agent halts.
	// This bounds cost and prevents runaway loops.
	MaxSteps int
	// MaxTokens is the per-call output token limit.
	MaxTokens int64
	// Timeout is the wall-clock limit for the entire run.
	// Zero means no timeout beyond the context deadline.
	Timeout time.Duration
	// SystemPrompt prepends custom instructions before any active skill instructions.
	// Defaults to a minimal task-completion prompt when empty.
	SystemPrompt string
	// Hooks receives lifecycle events from the agent loop. All fields are optional.
	Hooks Hooks
}

// ModelFromString converts a plain string to an anthropic.Model.
// Use this when reading the model name from a config file or CLI flag.
func ModelFromString(s string) anthropic.Model {
	return anthropic.Model(s)
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Model:     DefaultModel,
		MaxSteps:  DefaultMaxSteps,
		MaxTokens: DefaultMaxTokens,
	}
}
