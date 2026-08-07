// Package anthropic is a BYOK adapter for Anthropic's Messages API.
// Wave 0 only wires Ping; Chat/Embed/Transcribe ship in later waves.
package anthropic

import (
	"bytes"
	"context"
	"encoding/base64"
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

const defaultBaseURL = "https://api.anthropic.com"
const defaultChatModel = "claude-haiku-4-5"
const apiVersion = "2023-06-01"

type Client struct {
	apiKey    string
	baseURL   string
	chatModel string
	http      *http.Client
}

func New(apiKey, chatModel, baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if chatModel == "" {
		chatModel = defaultChatModel
	}
	return &Client{apiKey: apiKey, baseURL: baseURL, chatModel: chatModel, http: httpClient}
}

func (c *Client) Capabilities() []domain.AICapability {
	return []domain.AICapability{domain.AICapChat}
}

// Ping performs the cheapest possible Messages call (max_tokens=1) to validate
// the API key and chosen model.
func (c *Client) Ping(ctx context.Context) error {
	body := map[string]any{
		"model":      c.chatModel,
		"max_tokens": 1,
		"messages": []map[string]any{
			{"role": "user", "content": "ping"},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("content-type", "application/json")
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

// Chat issues a single Messages API call and emits one content event plus a
// final terminal event with usage counts. It does not stream intra-message
// tokens yet — feature-level latency is dominated by first-token delay,
// which is fine for F1/F4/F10. Streaming lands with F3 multi-file chat.
func (c *Client) Chat(ctx context.Context, req ports.ChatRequest) (<-chan ports.ChatEvent, error) {
	model := req.Model
	if model == "" {
		model = c.chatModel
	}
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	// Anthropic treats the system role as a top-level field rather than a
	// message entry. Extract it if present.
	var system string
	msgs := make([]map[string]any, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == "system" {
			if system != "" {
				system += "\n\n"
			}
			system += m.Content
			continue
		}
		msgs = append(msgs, map[string]any{"role": m.Role, "content": messageContent(m)})
	}

	body := map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		"messages":   msgs,
	}
	if system != "" {
		body["system"] = system
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)
	httpReq.Header.Set("content-type", "application/json")
	res, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, decodeAPIError(res)
	}

	var decoded struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("anthropic: decode response: %w", err)
	}
	var text string
	for _, part := range decoded.Content {
		if part.Type == "text" {
			text += part.Text
		}
	}

	ch := make(chan ports.ChatEvent, 2)
	if text != "" {
		ch <- ports.ChatEvent{Delta: text}
	}
	finish := "stop"
	switch decoded.StopReason {
	case "", "stop_sequence", "end_turn":
		finish = "stop"
	case "max_tokens":
		finish = "length"
	default:
		finish = decoded.StopReason
	}
	ch <- ports.ChatEvent{
		InputTokens:  decoded.Usage.InputTokens,
		OutputTokens: decoded.Usage.OutputTokens,
		FinishReason: finish,
	}
	close(ch)
	return ch, nil
}

// messageContent returns either a plain string (no attachments) or a list of
// content blocks (attachments + text) that the Messages API accepts. Images
// and PDFs are base64-encoded as "document"/"image" source blocks. The text
// always comes last so the model reads the attachment first.
func messageContent(m ports.ChatMessage) any {
	if len(m.Attachments) == 0 {
		return m.Content
	}
	blocks := make([]map[string]any, 0, len(m.Attachments)+1)
	for _, att := range m.Attachments {
		mt := strings.ToLower(strings.TrimSpace(att.MediaType))
		encoded := base64.StdEncoding.EncodeToString(att.Data)
		block := map[string]any{
			"source": map[string]any{
				"type":       "base64",
				"media_type": mt,
				"data":       encoded,
			},
		}
		switch {
		case mt == "application/pdf":
			block["type"] = "document"
		case strings.HasPrefix(mt, "image/"):
			block["type"] = "image"
		default:
			// Unknown media — skip rather than send a malformed block.
			continue
		}
		blocks = append(blocks, block)
	}
	if m.Content != "" {
		blocks = append(blocks, map[string]any{"type": "text", "text": m.Content})
	}
	if len(blocks) == 0 {
		return m.Content
	}
	return blocks
}

func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, int, error) {
	return nil, 0, errors.New("anthropic does not support embeddings")
}

func (c *Client) Transcribe(ctx context.Context, audio io.Reader, mime string) (string, error) {
	return "", errors.New("anthropic does not support transcription")
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type apiErrorEnvelope struct {
	Type  string    `json:"type"`
	Error *apiError `json:"error,omitempty"`
}

func decodeAPIError(res *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(res.Body, 8192))
	var env apiErrorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Error != nil {
		return fmt.Errorf("anthropic %d: %s: %s", res.StatusCode, env.Error.Type, env.Error.Message)
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = res.Status
	}
	return fmt.Errorf("anthropic %d: %s", res.StatusCode, msg)
}
