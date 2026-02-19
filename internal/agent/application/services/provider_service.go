package services

import (
	"fmt"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/infrastructure"
)

// ProviderService manages LLM provider construction with fallback support.
type ProviderService struct {
	cfg domain.LLMConfig
}

// NewProviderService creates a provider service from the given LLM config.
func NewProviderService(cfg domain.LLMConfig) *ProviderService {
	return &ProviderService{cfg: cfg}
}

// GetLLM returns a usable LLM, trying primary first then fallback.
// Returns error only if both providers fail (or no fallback configured).
func (s *ProviderService) GetLLM() (infrastructure.LLMModel, error) {
	llm, err := infrastructure.NewLLM(s.cfg.Primary)
	if err == nil {
		return llm, nil
	}

	if s.cfg.Fallback == nil {
		return nil, fmt.Errorf("provider_service: primary provider failed and no fallback configured: %w", err)
	}

	fallbackLLM, fallbackErr := infrastructure.NewLLM(*s.cfg.Fallback)
	if fallbackErr != nil {
		return nil, fmt.Errorf("provider_service: both primary (%v) and fallback (%v) providers failed", err, fallbackErr)
	}
	return fallbackLLM, nil
}
