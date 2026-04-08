package builtin

// SummarizeSkill instructs the agent to produce tight, structured summaries.
type SummarizeSkill struct{}

func NewSummarizeSkill() *SummarizeSkill { return &SummarizeSkill{} }

func (s *SummarizeSkill) Name() string { return "summarize" }

func (s *SummarizeSkill) Description() string {
	return "Produce concise, structured summaries of long content."
}

func (s *SummarizeSkill) Instructions() string {
	return `When summarizing content:
- Lead with the core finding or conclusion, not background.
- Use headers to separate distinct sections if the source has multiple topics.
- Bullet key points, decisions, and action items where applicable.
- Omit filler, repetition, and preamble.
- State explicitly if information is absent or ambiguous rather than filling gaps.
- Target the shortest summary that preserves all load-bearing facts.`
}
