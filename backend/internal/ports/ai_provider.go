package ports

import (
	"context"
	"io"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
)

// ChatAttachment is a binary payload sent alongside the text content of a
// user message. Providers that support vision/PDF input (e.g. Anthropic,
// OpenAI GPT-4o) serialize it as a structured content block; providers
// without vision ignore the slice.
type ChatAttachment struct {
	MediaType string // "application/pdf", "image/png", "image/jpeg", "image/webp", "image/gif"
	Data      []byte
}

// ChatMessage is a single turn in a chat completion request.
type ChatMessage struct {
	Role        string // "system" | "user" | "assistant"
	Content     string
	Attachments []ChatAttachment
}

// ChatRequest describes a chat completion. Streaming-first: the caller
// receives a channel of events, including the terminal one with token counts.
type ChatRequest struct {
	Model       string
	Messages    []ChatMessage
	MaxTokens   int
	Temperature float32
	// JSONMode asks the provider to produce strict JSON output when supported.
	JSONMode bool
}

// ChatEvent is either a content delta or a terminal result. The terminal
// event has a non-empty FinishReason ("stop", "length", "error"). On error
// the Err field is set; on success it is nil.
type ChatEvent struct {
	Delta        string
	InputTokens  int
	OutputTokens int
	FinishReason string
	Err          error
}

// AIProvider is the minimal surface every upstream LLM service must
// implement. Wave 0 only requires Ping() + Capabilities(); the rest is
// filled in as features land in later waves.
type AIProvider interface {
	// Chat streams tokens from the provider. The returned channel is closed
	// once the terminal event (or first error) is sent.
	Chat(ctx context.Context, req ChatRequest) (<-chan ChatEvent, error)

	// Embed returns a vector per input text plus the provider-reported input
	// token count.
	Embed(ctx context.Context, texts []string) ([][]float32, int, error)

	// Transcribe converts spoken audio to text. Providers that don't support
	// audio return an error.
	Transcribe(ctx context.Context, audio io.Reader, mime string) (string, error)

	// Ping is a cheap, low-latency validation call used when saving a config
	// or testing it from the UI. It must return quickly (< ~5s).
	Ping(ctx context.Context) error

	// Capabilities enumerates what this provider instance can actually do,
	// based on its configuration (model selection, declared caps).
	Capabilities() []domain.AICapability
}
