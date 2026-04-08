package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"

	"github.com/akshaymemane/superclaw/internal/skills"
	"github.com/akshaymemane/superclaw/internal/tools"
)

// Agent is the core runtime. It owns the execution loop, tool dispatch,
// and LLM interaction. One Agent per task run.
type Agent struct {
	config Config
	state  *State
	tools  *tools.Registry
	skills *skills.Registry
	llm    LLMClient
}

// New creates an Agent backed by the real Anthropic API.
// apiKey falls back to the ANTHROPIC_API_KEY environment variable when empty.
func New(cfg Config, t *tools.Registry, s *skills.Registry, apiKey string) *Agent {
	return &Agent{
		config: cfg,
		state:  newState(),
		tools:  t,
		skills: s,
		llm:    newRealLLM(apiKey),
	}
}

// NewWithLLM creates an Agent with a custom LLMClient.
// Use this in tests to inject a mock without hitting the real API.
func NewWithLLM(cfg Config, t *tools.Registry, s *skills.Registry, llm LLMClient) *Agent {
	return &Agent{
		config: cfg,
		state:  newState(),
		tools:  t,
		skills: s,
		llm:    llm,
	}
}

// Run executes the plan → act → observe loop for the given task.
// It returns the final State; the same pointer is accessible via State().
func (a *Agent) Run(ctx context.Context, task string) (*State, error) {
	if a.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.config.Timeout)
		defer cancel()
	}

	a.state.mu.Lock()
	a.state.Status = StatusRunning
	a.state.mu.Unlock()

	systemPrompt := a.buildSystemPrompt()
	toolParams := a.tools.Params()

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(task)),
	}

	for {
		if a.state.Steps >= a.config.MaxSteps {
			err := fmt.Errorf("step limit reached (%d)", a.config.MaxSteps)
			a.state.fail(err)
			return a.state, err
		}
		a.state.incrementStep()

		if a.config.Hooks.OnStep != nil {
			a.config.Hooks.OnStep(a.state.Steps)
		}

		params := anthropic.MessageNewParams{
			Model:     a.config.Model,
			MaxTokens: a.config.MaxTokens,
			System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
			Messages:  messages,
		}
		if len(toolParams) > 0 {
			params.Tools = toolParams
		}

		resp, err := a.llm.Create(ctx, params)
		if err != nil {
			a.state.fail(err)
			return a.state, err
		}

		// Append the full assistant turn (preserves tool_use blocks for the API).
		messages = append(messages, resp.ToParam())

		if resp.StopReason != anthropic.StopReasonToolUse {
			a.state.finish(extractText(resp.Content))
			return a.state, nil
		}

		// Execute every tool the model requested and collect results.
		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range resp.Content {
			toolUse, ok := block.AsAny().(anthropic.ToolUseBlock)
			if !ok {
				continue
			}

			rawInput := json.RawMessage(toolUse.JSON.Input.Raw())
			if a.config.Hooks.OnToolCall != nil {
				a.config.Hooks.OnToolCall(a.state.Steps, toolUse.Name, rawInput)
			}

			result, execErr := a.executeTool(ctx, toolUse)

			isError := execErr != nil
			resultStr := result
			if isError {
				resultStr = execErr.Error()
			}
			if a.config.Hooks.OnToolResult != nil {
				a.config.Hooks.OnToolResult(a.state.Steps, toolUse.Name, resultStr, isError)
			}

			if isError {
				toolResults = append(toolResults,
					anthropic.NewToolResultBlock(toolUse.ID, resultStr, true))
			} else {
				toolResults = append(toolResults,
					anthropic.NewToolResultBlock(toolUse.ID, result, false))
			}
		}

		if len(toolResults) > 0 {
			messages = append(messages, anthropic.NewUserMessage(toolResults...))
		}
	}
}

// State returns the current run state. Safe to call after Run returns.
func (a *Agent) State() *State { return a.state }

func (a *Agent) executeTool(ctx context.Context, toolUse anthropic.ToolUseBlock) (string, error) {
	t, err := a.tools.Get(toolUse.Name)
	if err != nil {
		return "", err
	}
	return t.Execute(ctx, json.RawMessage(toolUse.JSON.Input.Raw()))
}

func (a *Agent) buildSystemPrompt() string {
	base := a.config.SystemPrompt
	if base == "" {
		base = "You are a focused agent. Use the available tools to complete the given task. Stop when the task is done."
	}
	if instructions := a.skills.ActiveInstructions(); instructions != "" {
		base = base + "\n\n" + instructions
	}
	return base
}

// extractText concatenates all TextBlocks in a content slice.
// ThinkingBlocks are intentionally skipped.
func extractText(content []anthropic.ContentBlockUnion) string {
	var sb strings.Builder
	for _, block := range content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(t.Text)
		}
	}
	return sb.String()
}
