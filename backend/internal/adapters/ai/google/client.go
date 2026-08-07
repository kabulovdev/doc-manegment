// Package google is a BYOK adapter for Google Gemini (generativelanguage.googleapis.com).
// Supports chat + embed.
package google

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

const defaultBaseURL = "https://generativelanguage.googleapis.com"
const defaultChatModel = "gemini-2.0-flash"
const defaultEmbedModel = "text-embedding-004"

type Client struct {
	apiKey     string
	baseURL    string
	chatModel  string
	embedModel string
	http       *http.Client
}

func New(apiKey, chatModel, embedModel, baseURL string, httpClient *http.Client) *Client {
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
	if embedModel == "" {
		embedModel = defaultEmbedModel
	}
	return &Client{apiKey: apiKey, baseURL: baseURL, chatModel: chatModel, embedModel: embedModel, http: httpClient}
}

func (c *Client) Capabilities() []domain.AICapability {
	return []domain.AICapability{domain.AICapChat, domain.AICapEmbed}
}

// Ping lists Gemini models — cheap, authed, covers the common failure modes
// (bad key, wrong base URL, reachability).
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1beta/models?pageSize=1&key="+c.apiKey, nil)
	if err != nil {
		return err
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

type geminiContent struct {
	Role  string        `json:"role,omitempty"`
	Parts []geminiPart  `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

func (c *Client) Chat(ctx context.Context, req ports.ChatRequest) (<-chan ports.ChatEvent, error) {
	model := req.Model
	if model == "" {
		model = c.chatModel
	}
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	var system string
	contents := make([]geminiContent, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m.Role == "system" {
			if system != "" {
				system += "\n\n"
			}
			system += m.Content
			continue
		}
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{Role: role, Parts: []geminiPart{{Text: m.Content}}})
	}

	body := map[string]any{
		"contents": contents,
		"generationConfig": map[string]any{
			"maxOutputTokens": maxTokens,
			"temperature":     req.Temperature,
		},
	}
	if system != "" {
		body["systemInstruction"] = geminiContent{Parts: []geminiPart{{Text: system}}}
	}
	if req.JSONMode {
		body["generationConfig"].(map[string]any)["responseMimeType"] = "application/json"
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", c.baseURL, model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
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
		Candidates []struct {
			Content      geminiContent `json:"content"`
			FinishReason string        `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("google: decode response: %w", err)
	}
	var text, finish string
	if len(decoded.Candidates) > 0 {
		for _, part := range decoded.Candidates[0].Content.Parts {
			text += part.Text
		}
		finish = decoded.Candidates[0].FinishReason
	}
	ch := make(chan ports.ChatEvent, 2)
	if text != "" {
		ch <- ports.ChatEvent{Delta: text}
	}
	switch strings.ToUpper(finish) {
	case "", "STOP":
		finish = "stop"
	case "MAX_TOKENS":
		finish = "length"
	default:
		finish = strings.ToLower(finish)
	}
	ch <- ports.ChatEvent{
		InputTokens:  decoded.UsageMetadata.PromptTokenCount,
		OutputTokens: decoded.UsageMetadata.CandidatesTokenCount,
		FinishReason: finish,
	}
	close(ch)
	return ch, nil
}

// Embed uses batchEmbedContents for a single HTTP call regardless of input count.
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, int, error) {
	if len(texts) == 0 {
		return nil, 0, nil
	}
	reqs := make([]map[string]any, 0, len(texts))
	for _, t := range texts {
		reqs = append(reqs, map[string]any{
			"model":   "models/" + c.embedModel,
			"content": geminiContent{Parts: []geminiPart{{Text: t}}},
		})
	}
	body := map[string]any{"requests": reqs}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:batchEmbedContents?key=%s", c.baseURL, c.embedModel, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, 0, err
	}
	httpReq.Header.Set("content-type", "application/json")
	res, err := c.http.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, 0, decodeAPIError(res)
	}
	var decoded struct {
		Embeddings []struct {
			Values []float32 `json:"values"`
		} `json:"embeddings"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, 0, fmt.Errorf("google: decode embed response: %w", err)
	}
	out := make([][]float32, 0, len(decoded.Embeddings))
	for _, e := range decoded.Embeddings {
		out = append(out, e.Values)
	}
	// Gemini's batch endpoint doesn't return input token counts; approximate
	// as sum of character counts / 4 so cost tracking has something useful.
	approxTokens := 0
	for _, t := range texts {
		approxTokens += len(t) / 4
	}
	return out, approxTokens, nil
}

func (c *Client) Transcribe(ctx context.Context, audio io.Reader, mime string) (string, error) {
	return "", errors.New("google does not support audio transcription via this adapter")
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type apiErrorEnvelope struct {
	Error *apiError `json:"error,omitempty"`
}

func decodeAPIError(res *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(res.Body, 8192))
	var env apiErrorEnvelope
	if err := json.Unmarshal(body, &env); err == nil && env.Error != nil {
		return fmt.Errorf("google %d: %s: %s", res.StatusCode, env.Error.Status, env.Error.Message)
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = res.Status
	}
	return fmt.Errorf("google %d: %s", res.StatusCode, msg)
}
