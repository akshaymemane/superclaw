package tools

import (
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// Registry holds registered tools and an allowlist that controls
// which subset the agent may actually call.
type Registry struct {
	all       map[string]Tool
	allowlist map[string]struct{}
}

func NewRegistry() *Registry {
	return &Registry{
		all:       make(map[string]Tool),
		allowlist: make(map[string]struct{}),
	}
}

// Register adds a tool to the registry. It is not usable until Allow is called.
func (r *Registry) Register(t Tool) {
	r.all[t.Name()] = t
}

// Allow marks one or more tool names as usable by the agent.
func (r *Registry) Allow(names ...string) {
	for _, n := range names {
		r.allowlist[n] = struct{}{}
	}
}

// Get returns a tool only if it is both registered and allowed.
// The agent must never dispatch a tool that fails this check.
func (r *Registry) Get(name string) (Tool, error) {
	if _, ok := r.allowlist[name]; !ok {
		return nil, fmt.Errorf("tool %q is not in the allowlist", name)
	}
	t, ok := r.all[name]
	if !ok {
		return nil, fmt.Errorf("tool %q is not registered", name)
	}
	return t, nil
}

// Params returns Anthropic SDK tool params for all allowed, registered tools.
func (r *Registry) Params() []anthropic.ToolUnionParam {
	out := make([]anthropic.ToolUnionParam, 0, len(r.allowlist))
	for name := range r.allowlist {
		if t, ok := r.all[name]; ok {
			out = append(out, ToParam(t))
		}
	}
	return out
}
