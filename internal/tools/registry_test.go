package tools_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/akshaymemane/superclaw/internal/tools"
)

type echoTool struct{ name string }

func (e *echoTool) Name() string               { return e.name }
func (e *echoTool) Description() string        { return "echo" }
func (e *echoTool) Properties() map[string]any { return map[string]any{} }
func (e *echoTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	return string(input), nil
}

func TestRegistry_GetRequiresAllowlist(t *testing.T) {
	r := tools.NewRegistry()
	r.Register(&echoTool{name: "echo"})

	if _, err := r.Get("echo"); err == nil {
		t.Fatal("expected error: tool registered but not allowed")
	}

	r.Allow("echo")
	if _, err := r.Get("echo"); err != nil {
		t.Fatalf("unexpected error after Allow: %v", err)
	}
}

func TestRegistry_GetUnregistered(t *testing.T) {
	r := tools.NewRegistry()
	r.Allow("ghost")
	if _, err := r.Get("ghost"); err == nil {
		t.Fatal("expected error: tool allowed but not registered")
	}
}

func TestRegistry_ParamsOnlyAllowed(t *testing.T) {
	r := tools.NewRegistry()
	r.Register(&echoTool{name: "a"})
	r.Register(&echoTool{name: "b"})
	r.Allow("a")

	params := r.Params()
	if len(params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(params))
	}
}
