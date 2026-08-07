// Package openai is a BYOK adapter that speaks the OpenAI HTTP protocol.
// The same adapter also serves OpenAI-compatible providers (OpenRouter, Groq,
// Together, vLLM, LM Studio) and local Ollama — those just set BaseURL.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
)

const defaultBaseURL = "https://api.openai.com"
const defaultOllamaBaseURL = "http://ollama:11434"

type Config struct {
	APIKey          string
	BaseURL         string
	ChatModel       string
	EmbedModel      string
	TranscribeModel string
	Capabilities    []domain.AICapability
	Kind            domain.AIProviderKind
}

type Client struct {
	cfg  Config
	http *http.Client
}

func New(cfg Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if cfg.BaseURL == "" {
		if cfg.Kind == domain.AIProviderOllama {
			cfg.BaseURL = defaultOllamaBaseURL
		} else {
			cfg.BaseURL = defaultBaseURL
		}
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	return &Client{cfg: cfg, http: httpClient}
}

func (c *Client) Capabilities() []domain.AICapability {
	if len(c.cfg.Capabilities) > 0 {
		return c.cfg.Capabilities
	}
	// Sensible defaults per provider kind.
	switch c.cfg.Kind {
	case domain.AIProviderOllama:
		return []domain.AICapability{domain.AICapChat, domain.AICapEmbed}
	case domain.AIProviderOpenAICompat:
		return []domain.AICapability{domain.AICapChat}
	}
	return []domain.AICapability{domain.AICapChat, domain.AICapEmbed, domain.AICapTranscribe}
}

// Ping lists available models — cheapest validation that hits the key and the
// base URL simultaneously. Ollama also supports GET /api/tags, but /v1/models
// works for Ollama's OpenAI-compatibility mode.
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+"/v1/models", nil)
	if err != nil {
		return err
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return decodeAPIError(res)
}

// Chat does a non-streaming /v1/chat/completions call and emits a single
// content event plus a terminal event with usage counts. Streaming lands
// with F3 multi-file chat.
func (c *Client) Chat(ctx context.Context, req ports.ChatRequest) (<-chan ports.ChatEvent, error) {
	model := req.Model
	if model == "" {
		model = c.cfg.ChatModel
	}
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}
	msgs := make([]map[string]any, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]any{"role": m.Role, "content": m.Content})
	}
	body := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		"messages":   msgs,
		"stream":     false,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.JSONMode {
		body["response_format"] = map[string]string{"type": "json_object"}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("content-type", "application/json")
	if c.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	res, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, decodeAPIError(res)
	}

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("openai: decode response: %w", err)
	}
	text := ""
	finish := "stop"
	if len(decoded.Choices) > 0 {
		text = decoded.Choices[0].Message.Content
		if decoded.Choices[0].FinishReason != "" {
			finish = decoded.Choices[0].FinishReason
		}
	}
	ch := make(chan ports.ChatEvent, 2)
	if text != "" {
		ch <- ports.ChatEvent{Delta: text}
	}
	ch <- ports.ChatEvent{
		InputTokens:  decoded.Usage.PromptTokens,
		OutputTokens: decoded.Usage.CompletionTokens,
		FinishReason: finish,
	}
	close(ch)
	return ch, nil
}

// Embed calls /v1/embeddings and returns one vector per input plus the total
// input token count reported by the provider.
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, int, error) {
	if len(texts) == 0 {
		return nil, 0, nil
	}
	model := c.cfg.EmbedModel
	if model == "" {
		model = "text-embedding-3-small"
	}
	body := map[string]any{
		"model": model,
		"input": texts,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/v1/embeddings", bytes.NewReader(raw))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("content-type", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, 0, decodeAPIError(res)
	}
	var decoded struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, 0, fmt.Errorf("openai: decode embed response: %w", err)
	}
	out := make([][]float32, len(decoded.Data))
	for _, d := range decoded.Data {
		if d.Index < 0 || d.Index >= len(out) {
			continue
		}
		out[d.Index] = d.Embedding
	}
	return out, decoded.Usage.PromptTokens, nil
}

func (c *Client) Transcribe(ctx context.Context, audio io.Reader, mime string) (string, error) {
	return "", errors.New("openai transcribe: not yet implemented (Wave 5)")
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type apiErrorEnvelope struct {
	Error *apiError `json:"error,omitempty"`
}

func decodeAPIError(res *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(res.Body, 8192))
	var env apiErrorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Error != nil {
		return fmt.Errorf("openai %d: %s", res.StatusCode, env.Error.Message)
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = res.Status
	}
	return fmt.Errorf("openai %d: %s", res.StatusCode, msg)
}
