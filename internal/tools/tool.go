package tools

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
)

// Tool is the interface every superclaw tool must implement.
// Keep the set of registered tools small and auditable — 3-6 tools is ideal.
type Tool interface {
	Name() string
	Description() string
	// Properties returns a JSON-schema-compatible properties map describing
	// the tool's input fields. Used to build the Anthropic tool schema.
	Properties() map[string]any
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}

// ToParam converts a Tool to the Anthropic SDK ToolUnionParam.
func ToParam(t Tool) anthropic.ToolUnionParam {
	param := anthropic.ToolParam{
		Name:        t.Name(),
		Description: anthropic.String(t.Description()),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: t.Properties(),
		},
	}
	return anthropic.ToolUnionParam{OfTool: &param}
}
