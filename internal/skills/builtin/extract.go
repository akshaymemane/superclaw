package builtin

// ExtractSkill instructs the agent to pull structured data from unstructured
// content with high fidelity and explicit handling of missing fields.
type ExtractSkill struct{}

func NewExtractSkill() *ExtractSkill { return &ExtractSkill{} }

func (e *ExtractSkill) Name() string { return "extract" }

func (e *ExtractSkill) Description() string {
	return "Extract structured data from unstructured content with explicit missing-field handling."
}

func (e *ExtractSkill) Instructions() string {
	return `When extracting structured data:
1. Identify the target schema before extracting — list the fields and their expected types.
2. Extract each field independently; do not infer or guess values not present in the source.
3. For missing or ambiguous fields, output null or a clearly labelled placeholder such as "NOT_FOUND".
4. Preserve the original casing and formatting of values unless normalization is explicitly requested.
5. If the source contains multiple records, extract all of them in a consistent format.
6. When the source is a file, read it first before attempting extraction.
7. Return extracted data as JSON or a clearly structured list.`
}
