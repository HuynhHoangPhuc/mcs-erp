package domain

// LLMProvider identifies which AI provider to use.
type LLMProvider string

const (
	ProviderClaude LLMProvider = "claude"
	ProviderOpenAI LLMProvider = "openai"
	ProviderOllama LLMProvider = "ollama"
)

// ProviderConfig holds the configuration for a single LLM provider.
type ProviderConfig struct {
	Provider LLMProvider
	Model    string
	APIKey   string
	BaseURL  string // used for Ollama server URL
}

// LLMConfig holds the primary and optional fallback provider configurations.
type LLMConfig struct {
	Primary  ProviderConfig
	Fallback *ProviderConfig // nil means no fallback
}
