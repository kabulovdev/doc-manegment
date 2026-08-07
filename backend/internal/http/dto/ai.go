package dto

import (
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

type CreateAIConfigRequest struct {
	DisplayName     string   `json:"display_name" validate:"required,min=1,max=100"`
	Provider        string   `json:"provider" validate:"required"`
	BaseURL         string   `json:"base_url,omitempty"`
	APIKey          string   `json:"api_key"`
	ChatModel       string   `json:"chat_model,omitempty"`
	EmbedModel      string   `json:"embed_model,omitempty"`
	TranscribeModel string   `json:"transcribe_model,omitempty"`
	Capabilities    []string `json:"capabilities,omitempty"`
	DefaultChat     bool     `json:"default_chat,omitempty"`
	DefaultEmbed    bool     `json:"default_embed,omitempty"`
	DefaultTrans    bool     `json:"default_transcribe,omitempty"`
}

type UpdateAIConfigRequest struct {
	DisplayName     *string   `json:"display_name,omitempty" validate:"omitempty,min=1,max=100"`
	ChatModel       *string   `json:"chat_model,omitempty"`
	EmbedModel      *string   `json:"embed_model,omitempty"`
	TranscribeModel *string   `json:"transcribe_model,omitempty"`
	Capabilities    *[]string `json:"capabilities,omitempty"`
}

type SetDefaultRequest struct {
	Capability string `json:"capability" validate:"required,oneof=chat embed transcribe"`
}

type AIConfigResponse struct {
	ID                  string     `json:"id"`
	DisplayName         string     `json:"display_name"`
	Provider            string     `json:"provider"`
	BaseURL             string     `json:"base_url,omitempty"`
	KeyPrefix           string     `json:"key_prefix,omitempty"`
	ChatModel           string     `json:"chat_model,omitempty"`
	EmbedModel          string     `json:"embed_model,omitempty"`
	TranscribeModel     string     `json:"transcribe_model,omitempty"`
	Capabilities        []string   `json:"capabilities"`
	IsDefaultChat       bool       `json:"is_default_chat"`
	IsDefaultEmbed      bool       `json:"is_default_embed"`
	IsDefaultTranscribe bool       `json:"is_default_transcribe"`
	UsedTokensIn        int64      `json:"used_tokens_in"`
	UsedTokensOut       int64      `json:"used_tokens_out"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty"`
	LastTestedAt        *time.Time `json:"last_tested_at,omitempty"`
	LastError           string     `json:"last_error,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

type AIUsageTotalsResponse struct {
	AIConfigID string `json:"ai_config_id"`
	TokensIn   int64  `json:"tokens_in"`
	TokensOut  int64  `json:"tokens_out"`
	Calls      int64  `json:"calls"`
}

// SuggestedFieldDTO is F1's per-row output.
type SuggestedFieldDTO struct {
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	Type       string  `json:"type"`
	Confidence float32 `json:"confidence,omitempty"`
}

// SuggestedTagDTO is F4's per-row output.
type SuggestedTagDTO struct {
	Name       string  `json:"name"`
	Confidence float32 `json:"confidence,omitempty"`
}

// AISuggestionsMeta carries provider/model/token info alongside suggestions.
type AISuggestionsMeta struct {
	Provider         string `json:"provider,omitempty"`
	Model            string `json:"model,omitempty"`
	TokensIn         int    `json:"tokens_in,omitempty"`
	TokensOut        int    `json:"tokens_out,omitempty"`
	LatencyMS        int64  `json:"latency_ms,omitempty"`
	ExtractionStatus string `json:"extraction_status"`
}

type SuggestFieldsResponse struct {
	Fields []SuggestedFieldDTO `json:"fields"`
	Meta   AISuggestionsMeta   `json:"meta"`
}

type SuggestTagsResponse struct {
	Tags []SuggestedTagDTO `json:"tags"`
	Meta AISuggestionsMeta `json:"meta"`
}

type SuggestNameResponse struct {
	Name string            `json:"name"`
	Meta AISuggestionsMeta `json:"meta"`
}

// Semantic search (F2)
type SemanticSearchRequest struct {
	Query string `json:"query" validate:"required,min=1,max=1000"`
	Limit int    `json:"limit,omitempty"`
}

type SemanticSearchHit struct {
	FileID      string  `json:"file_id"`
	Score       float32 `json:"score"`
	Name        string  `json:"name"`
	Mime        string  `json:"mime,omitempty"`
	FolderID    string  `json:"folder_id,omitempty"`
	TextPreview string  `json:"text_preview,omitempty"`
}

type SemanticSearchResponse struct {
	Hits []SemanticSearchHit `json:"hits"`
}

// Ask your docs (F3)
type AskRequest struct {
	Question string `json:"question" validate:"required,min=1,max=1000"`
	Limit    int    `json:"limit,omitempty"`
}

// FileChatRequest is the payload for per-file chat (no RAG).
type FileChatRequest struct {
	Question string `json:"question" validate:"required,min=1,max=2000"`
}

type AskCitationDTO struct {
	Index   int     `json:"index"`
	FileID  string  `json:"file_id"`
	Name    string  `json:"name"`
	Mime    string  `json:"mime,omitempty"`
	Score   float32 `json:"score,omitempty"`
	Snippet string  `json:"snippet,omitempty"`
}

type AskResponse struct {
	Answer    string           `json:"answer"`
	Citations []AskCitationDTO `json:"citations"`
	Provider  string           `json:"provider,omitempty"`
	Model     string           `json:"model,omitempty"`
	TokensIn  int              `json:"tokens_in,omitempty"`
	TokensOut int              `json:"tokens_out,omitempty"`
	LatencyMS int64            `json:"latency_ms,omitempty"`
}

// Suggest folder (F5)
type SuggestFolderResponse struct {
	FolderID   string  `json:"folder_id,omitempty"`
	FolderName string  `json:"folder_name,omitempty"`
	Score      float32 `json:"score,omitempty"`
}

// ProcessDocumentRequest kicks off the vision-AI extraction for a file.
// ProviderID is optional — empty means "use my default chat provider".
type ProcessDocumentRequest struct {
	ProviderID string `json:"provider_id,omitempty"`
}

// AIExtractionResponse mirrors domain.AIExtraction for the frontend.
// Status values: "", "pending", "ready", "failed".
type AIExtractionResponse struct {
	Status      string     `json:"status"`
	Text        string     `json:"text,omitempty"`
	Model       string     `json:"model,omitempty"`
	Provider    string     `json:"provider,omitempty"`
	ProviderID  string     `json:"provider_id,omitempty"`
	TokensIn    int        `json:"tokens_in,omitempty"`
	TokensOut   int        `json:"tokens_out,omitempty"`
	ExtractedAt *time.Time `json:"extracted_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

func AIExtractionToResponse(e *domain.AIExtraction) AIExtractionResponse {
	status := string(e.Status)
	if status == "" {
		status = "none"
	}
	providerID := ""
	if !e.ProviderID.IsZero() {
		providerID = e.ProviderID.Hex()
	}
	return AIExtractionResponse{
		Status:      status,
		Text:        e.Text,
		Model:       e.Model,
		Provider:    e.Provider,
		ProviderID:  providerID,
		TokensIn:    e.TokensIn,
		TokensOut:   e.TokensOut,
		ExtractedAt: e.ExtractedAt,
		Error:       e.Error,
	}
}

func AIConfigToResponse(c *domain.AIConfig) AIConfigResponse {
	caps := make([]string, 0, len(c.Capabilities))
	for _, k := range c.Capabilities {
		caps = append(caps, string(k))
	}
	return AIConfigResponse{
		ID:                  c.ID.Hex(),
		DisplayName:         c.DisplayName,
		Provider:            string(c.Provider),
		BaseURL:             c.BaseURL,
		KeyPrefix:           c.KeyPrefix,
		ChatModel:           c.ChatModel,
		EmbedModel:          c.EmbedModel,
		TranscribeModel:     c.TranscribeModel,
		Capabilities:        caps,
		IsDefaultChat:       c.IsDefaultChat,
		IsDefaultEmbed:      c.IsDefaultEmbed,
		IsDefaultTranscribe: c.IsDefaultTranscribe,
		UsedTokensIn:        c.UsedTokensIn,
		UsedTokensOut:       c.UsedTokensOut,
		LastUsedAt:          c.LastUsedAt,
		LastTestedAt:        c.LastTestedAt,
		LastError:           c.LastError,
		CreatedAt:           c.CreatedAt,
	}
}
