package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AIProviderKind is the family of an upstream LLM/embedding/transcription
// service. Users bring their own keys; we never proxy through our account.
type AIProviderKind string

const (
	AIProviderAnthropic    AIProviderKind = "anthropic"
	AIProviderOpenAI       AIProviderKind = "openai"
	AIProviderGoogle       AIProviderKind = "google"
	AIProviderOpenAICompat AIProviderKind = "openai-compatible"
	AIProviderOllama       AIProviderKind = "ollama"
)

func (p AIProviderKind) Valid() bool {
	switch p {
	case AIProviderAnthropic, AIProviderOpenAI, AIProviderGoogle,
		AIProviderOpenAICompat, AIProviderOllama:
		return true
	}
	return false
}

// AICapability is a feature a provider may expose. A single AIConfig can
// declare multiple capabilities.
type AICapability string

const (
	AICapChat       AICapability = "chat"
	AICapEmbed      AICapability = "embed"
	AICapTranscribe AICapability = "transcribe"
)

func (c AICapability) Valid() bool {
	switch c {
	case AICapChat, AICapEmbed, AICapTranscribe:
		return true
	}
	return false
}

// AIConfig mirrors StorageConfig — per-user, encrypted credentials for a
// chosen AI provider. The plaintext API key never leaves the backend; only
// the masked prefix is exposed to the UI.
type AIConfig struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id"`
	DisplayName string             `bson:"display_name"`
	Provider    AIProviderKind     `bson:"provider"`
	BaseURL     string             `bson:"base_url,omitempty"`

	// Encrypted API key. Only persisted ciphertext; plaintext discarded after
	// adapter construction.
	APIKeyCT    []byte `bson:"api_key_ct"`
	APIKeyNonce []byte `bson:"api_key_nonce"`
	KeyVersion  int    `bson:"key_version"`
	// KeyPrefix is the first few characters of the plaintext key (e.g. "sk-…"),
	// persisted so the UI can display a masked identifier without decrypting.
	KeyPrefix string `bson:"key_prefix,omitempty"`

	ChatModel       string `bson:"chat_model,omitempty"`
	EmbedModel      string `bson:"embed_model,omitempty"`
	TranscribeModel string `bson:"transcribe_model,omitempty"`

	Capabilities []AICapability `bson:"capabilities"`

	IsDefaultChat       bool `bson:"is_default_chat,omitempty"`
	IsDefaultEmbed      bool `bson:"is_default_embed,omitempty"`
	IsDefaultTranscribe bool `bson:"is_default_transcribe,omitempty"`

	UsedTokensIn  int64      `bson:"used_tokens_in"`
	UsedTokensOut int64      `bson:"used_tokens_out"`
	LastUsedAt    *time.Time `bson:"last_used_at,omitempty"`
	LastError     string     `bson:"last_error,omitempty"`
	LastTestedAt  *time.Time `bson:"last_tested_at,omitempty"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// AIUsage is a row-per-request ledger. Tokens come from the provider's
// response metadata; never the prompt or completion content itself.
type AIUsage struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id"`
	AIConfigID  primitive.ObjectID `bson:"ai_config_id"`
	Feature     string             `bson:"feature"`
	Model       string             `bson:"model"`
	TokensIn    int                `bson:"tokens_in"`
	TokensOut   int                `bson:"tokens_out"`
	LatencyMS   int64              `bson:"latency_ms,omitempty"`
	Status      int                `bson:"status,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
}
