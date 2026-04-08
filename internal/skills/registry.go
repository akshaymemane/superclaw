package skills

import (
	"fmt"
	"strings"
)

// Registry holds registered skills and tracks which are active.
type Registry struct {
	all    map[string]Skill
	active map[string]struct{}
}

func NewRegistry() *Registry {
	return &Registry{
		all:    make(map[string]Skill),
		active: make(map[string]struct{}),
	}
}

// Register adds a skill to the registry. It is inactive until Activate is called.
func (r *Registry) Register(s Skill) {
	r.all[s.Name()] = s
}

// Activate marks skills as usable. Returns an error for any unknown name.
func (r *Registry) Activate(names ...string) error {
	for _, n := range names {
		if _, ok := r.all[n]; !ok {
			return fmt.Errorf("skill %q is not registered", n)
		}
		r.active[n] = struct{}{}
	}
	return nil
}

// ActiveInstructions returns the concatenated Instructions() of all active
// skills. Returns an empty string when no skills are active.
func (r *Registry) ActiveInstructions() string {
	if len(r.active) == 0 {
		return ""
	}
	var sb strings.Builder
	for name := range r.active {
		s := r.all[name]
		fmt.Fprintf(&sb, "## Skill: %s\n%s\n\n", s.Name(), s.Instructions())
	}
	return strings.TrimSpace(sb.String())
}
