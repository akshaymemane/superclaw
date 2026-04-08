package agent

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// LLMClient is the narrow interface the agent uses to call the language model.
// Implement this interface to use a different backend or a test double.
type LLMClient interface {
	Create(ctx context.Context, params anthropic.MessageNewParams) (anthropic.Message, error)
}

// realLLM wraps the Anthropic SDK client.
type realLLM struct {
	c anthropic.Client
}

func newRealLLM(apiKey string) LLMClient {
	var opts []option.RequestOption
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}
	return withRetry(&realLLM{c: anthropic.NewClient(opts...)})
}

func (r *realLLM) Create(ctx context.Context, params anthropic.MessageNewParams) (anthropic.Message, error) {
	resp, err := r.c.Messages.New(ctx, params)
	if err != nil {
		return anthropic.Message{}, err
	}
	return *resp, nil
}
