package infrastructure

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
)

// LLMModel is a minimal interface for calling an LLM; satisfied by langchaingo llms.Model.
type LLMModel interface {
	GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error)
}

// NewLLM constructs an LLM client for the given provider config.
// Supported providers: claude, openai, ollama.
func NewLLM(cfg domain.ProviderConfig) (LLMModel, error) {
	switch cfg.Provider {
	case domain.ProviderClaude:
		opts := []anthropic.Option{anthropic.WithModel(cfg.Model)}
		if cfg.APIKey != "" {
			opts = append(opts, anthropic.WithToken(cfg.APIKey))
		}
		return anthropic.New(opts...)

	case domain.ProviderOpenAI:
		opts := []openai.Option{openai.WithModel(cfg.Model)}
		if cfg.APIKey != "" {
			opts = append(opts, openai.WithToken(cfg.APIKey))
		}
		return openai.New(opts...)

	case domain.ProviderOllama:
		opts := []ollama.Option{ollama.WithModel(cfg.Model)}
		if cfg.BaseURL != "" {
			opts = append(opts, ollama.WithServerURL(cfg.BaseURL))
		}
		return ollama.New(opts...)

	default:
		return nil, fmt.Errorf("langchain_provider: unknown provider %q", cfg.Provider)
	}
}

// NewLLMWithFallback attempts to build the primary LLM; falls back to secondary on error.
func NewLLMWithFallback(primary, fallback domain.ProviderConfig) (LLMModel, error) {
	llm, err := NewLLM(primary)
	if err == nil {
		return llm, nil
	}
	return NewLLM(fallback)
}
