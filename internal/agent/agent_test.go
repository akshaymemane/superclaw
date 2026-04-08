package agent_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"

	"github.com/akshaymemane/superclaw/internal/agent"
	"github.com/akshaymemane/superclaw/internal/skills"
	"github.com/akshaymemane/superclaw/internal/tools"
)

// ── mock LLM ─────────────────────────────────────────────────────────────────

// mockLLM is a test double that returns pre-queued messages.
type mockLLM struct {
	queue []anthropic.Message
	idx   int
	// Calls records every set of params passed to Create, for inspection.
	Calls []anthropic.MessageNewParams
}

func (m *mockLLM) Create(_ context.Context, params anthropic.MessageNewParams) (anthropic.Message, error) {
	m.Calls = append(m.Calls, params)
	if m.idx >= len(m.queue) {
		return anthropic.Message{}, fmt.Errorf("mockLLM: queue exhausted after %d calls", m.idx)
	}
	msg := m.queue[m.idx]
	m.idx++
	return msg, nil
}

// ── message constructors ──────────────────────────────────────────────────────

func textMessage(text string) anthropic.Message {
	raw := fmt.Sprintf(
		`{"id":"msg_test","type":"message","role":"assistant","model":"claude-opus-4-6",`+
			`"stop_reason":"end_turn","stop_sequence":null,`+
			`"usage":{"input_tokens":10,"output_tokens":5},`+
			`"content":[{"type":"text","text":%s}]}`,
		jsonStr(text),
	)
	return mustUnmarshalMessage(raw)
}

func toolUseMessage(id, name string, input map[string]any) anthropic.Message {
	inputJSON, _ := json.Marshal(input)
	raw := fmt.Sprintf(
		`{"id":"msg_test","type":"message","role":"assistant","model":"claude-opus-4-6",`+
			`"stop_reason":"tool_use","stop_sequence":null,`+
			`"usage":{"input_tokens":10,"output_tokens":5},`+
			`"content":[{"type":"tool_use","id":%s,"name":%s,"input":%s}]}`,
		jsonStr(id), jsonStr(name), string(inputJSON),
	)
	return mustUnmarshalMessage(raw)
}

func mustUnmarshalMessage(raw string) anthropic.Message {
	var m anthropic.Message
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		panic(fmt.Sprintf("mustUnmarshalMessage: %v\nraw: %s", err, raw))
	}
	return m
}

func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// ── echo tool test double ─────────────────────────────────────────────────────

type echoTool struct{}

func (e *echoTool) Name() string               { return "echo" }
func (e *echoTool) Description() string        { return "echoes input back" }
func (e *echoTool) Properties() map[string]any { return map[string]any{"msg": map[string]any{"type": "string"}} }
func (e *echoTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	return string(input), nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestAgent(cfg agent.Config, llm agent.LLMClient, tr *tools.Registry) *agent.Agent {
	return agent.NewWithLLM(cfg, tr, skills.NewRegistry(), llm)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestAgent_EndTurn(t *testing.T) {
	mock := &mockLLM{queue: []anthropic.Message{textMessage("hello world")}}
	cfg := agent.DefaultConfig()

	a := newTestAgent(cfg, mock, tools.NewRegistry())
	state, err := a.Run(context.Background(), "say hello")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Status != agent.StatusDone {
		t.Errorf("status: want done, got %v", state.Status)
	}
	if state.Result != "hello world" {
		t.Errorf("result: want %q, got %q", "hello world", state.Result)
	}
	if state.Steps != 1 {
		t.Errorf("steps: want 1, got %d", state.Steps)
	}
	if len(mock.Calls) != 1 {
		t.Errorf("LLM calls: want 1, got %d", len(mock.Calls))
	}
}

func TestAgent_StepLimit(t *testing.T) {
	const limit = 3

	// Fill queue with enough tool-use responses to exceed the limit.
	queue := make([]anthropic.Message, limit+2)
	for i := range queue {
		queue[i] = toolUseMessage("tu1", "echo", map[string]any{"msg": "ping"})
	}
	mock := &mockLLM{queue: queue}

	cfg := agent.DefaultConfig()
	cfg.MaxSteps = limit

	tr := tools.NewRegistry()
	tr.Register(&echoTool{})
	tr.Allow("echo")

	a := newTestAgent(cfg, mock, tr)
	state, err := a.Run(context.Background(), "loop")

	if err == nil {
		t.Fatal("expected step-limit error, got nil")
	}
	if state.Status != agent.StatusError {
		t.Errorf("status: want error, got %v", state.Status)
	}
	if state.Steps != limit {
		t.Errorf("steps: want %d, got %d", limit, state.Steps)
	}
}

func TestAgent_ToolDispatch(t *testing.T) {
	mock := &mockLLM{queue: []anthropic.Message{
		toolUseMessage("tu1", "echo", map[string]any{"msg": "hi"}),
		textMessage("tool result received"),
	}}

	tr := tools.NewRegistry()
	tr.Register(&echoTool{})
	tr.Allow("echo")

	a := newTestAgent(agent.DefaultConfig(), mock, tr)
	state, err := a.Run(context.Background(), "use echo tool")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Result != "tool result received" {
		t.Errorf("result: want %q, got %q", "tool result received", state.Result)
	}
	if len(mock.Calls) != 2 {
		t.Errorf("LLM calls: want 2, got %d", len(mock.Calls))
	}
}

func TestAgent_DisallowedTool(t *testing.T) {
	// Model requests a tool that is not in the allowlist.
	mock := &mockLLM{queue: []anthropic.Message{
		toolUseMessage("tu1", "not_allowed", map[string]any{}),
		textMessage("ok"),
	}}

	tr := tools.NewRegistry()
	// "not_allowed" is never registered or allowed.

	a := newTestAgent(agent.DefaultConfig(), mock, tr)
	state, err := a.Run(context.Background(), "task")

	// The agent should not error out — it should report the tool error back to
	// the LLM and let the LLM respond with end_turn.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Result != "ok" {
		t.Errorf("result: want %q, got %q", "ok", state.Result)
	}
	// Two LLM calls: first returns tool_use, second receives the error result.
	if len(mock.Calls) != 2 {
		t.Errorf("LLM calls: want 2, got %d", len(mock.Calls))
	}
}

func TestAgent_HooksFired(t *testing.T) {
	mock := &mockLLM{queue: []anthropic.Message{
		toolUseMessage("tu1", "echo", map[string]any{"msg": "x"}),
		textMessage("done"),
	}}

	tr := tools.NewRegistry()
	tr.Register(&echoTool{})
	tr.Allow("echo")

	var (
		stepsCalled      []int
		toolCallsFired   []string
		toolResultsFired []string
	)

	cfg := agent.DefaultConfig()
	cfg.Hooks = agent.Hooks{
		OnStep: func(step int) {
			stepsCalled = append(stepsCalled, step)
		},
		OnToolCall: func(step int, name string, _ json.RawMessage) {
			toolCallsFired = append(toolCallsFired, name)
		},
		OnToolResult: func(step int, name string, result string, isError bool) {
			toolResultsFired = append(toolResultsFired, name)
		},
	}

	a := newTestAgent(cfg, mock, tr)
	if _, err := a.Run(context.Background(), "task"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stepsCalled) != 2 {
		t.Errorf("OnStep calls: want 2, got %d", len(stepsCalled))
	}
	if len(toolCallsFired) != 1 || toolCallsFired[0] != "echo" {
		t.Errorf("OnToolCall: want [echo], got %v", toolCallsFired)
	}
	if len(toolResultsFired) != 1 || toolResultsFired[0] != "echo" {
		t.Errorf("OnToolResult: want [echo], got %v", toolResultsFired)
	}
}
