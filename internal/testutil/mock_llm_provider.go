package testutil

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// MockLLMProvider is a deterministic LLM model for agent integration tests.
type MockLLMProvider struct {
	Response  string
	ToolCalls []llms.ToolCall
}

// GenerateContent implements agent/infrastructure.LLMModel.
func (m *MockLLMProvider) GenerateContent(ctx context.Context, _ []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	var callOpts llms.CallOptions
	for _, opt := range options {
		opt(&callOpts)
	}

	if callOpts.StreamingFunc != nil {
		if err := callOpts.StreamingFunc(ctx, []byte(m.Response)); err != nil {
			return nil, err
		}
	}

	choice := &llms.ContentChoice{
		Content:   m.Response,
		ToolCalls: m.ToolCalls,
	}

	if len(m.ToolCalls) > 0 {
		choice.FuncCall = m.ToolCalls[0].FunctionCall
	}

	return &llms.ContentResponse{Choices: []*llms.ContentChoice{choice}}, nil
}
