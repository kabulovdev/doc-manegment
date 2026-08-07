package ai

import (
	"fmt"
	"net/http"

	"gitlab.com/docuemnt_manegment/backend/internal/adapters/ai/anthropic"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/ai/google"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/ai/openai"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
)

// NewProvider constructs a concrete ports.AIProvider from a stored AIConfig +
// its decrypted API key. The httpClient is expected to carry an SSRF-guarded
// transport when baseURL is user-supplied.
func NewProvider(cfg *domain.AIConfig, apiKey string, httpClient *http.Client) (ports.AIProvider, error) {
	switch cfg.Provider {
	case domain.AIProviderAnthropic:
		return anthropic.New(apiKey, cfg.ChatModel, cfg.BaseURL, httpClient), nil
	case domain.AIProviderOpenAI, domain.AIProviderOpenAICompat, domain.AIProviderOllama:
		return openai.New(openai.Config{
			APIKey:          apiKey,
			BaseURL:         cfg.BaseURL,
			ChatModel:       cfg.ChatModel,
			EmbedModel:      cfg.EmbedModel,
			TranscribeModel: cfg.TranscribeModel,
			Capabilities:    cfg.Capabilities,
			Kind:            cfg.Provider,
		}, httpClient), nil
	case domain.AIProviderGoogle:
		return google.New(apiKey, cfg.ChatModel, cfg.EmbedModel, cfg.BaseURL, httpClient), nil
	}
	return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
}
