package skills

// Skill is a named capability that injects focused instructions into the
// agent's system prompt when active. Keep active skills to a small, curated
// set — each one should cover a single well-defined concern.
type Skill interface {
	Name() string
	// Description is a short one-line summary used for logging/introspection.
	Description() string
	// Instructions is appended to the system prompt when the skill is active.
	Instructions() string
}
