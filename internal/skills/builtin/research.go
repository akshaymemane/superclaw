package builtin

// ResearchSkill instructs the agent to gather and synthesize information
// from the web using the fetch_url tool.
type ResearchSkill struct{}

func NewResearchSkill() *ResearchSkill { return &ResearchSkill{} }

func (r *ResearchSkill) Name() string { return "research" }

func (r *ResearchSkill) Description() string {
	return "Gather and synthesize information from URLs using fetch_url."
}

func (r *ResearchSkill) Instructions() string {
	return `When researching a topic:
1. Identify 2-4 specific URLs likely to contain authoritative information.
2. Fetch each URL with fetch_url and extract the relevant content.
3. Cross-reference facts that appear in multiple sources.
4. Synthesize findings into a clear report with inline citations (URL + brief label).
5. Note conflicting information explicitly rather than picking one version silently.
6. If a URL fails or returns irrelevant content, try an alternative and note the failure.
7. Stop fetching once you have enough evidence; do not over-fetch.`
}
