// Package qdrant is a thin HTTP client for the Qdrant vector DB. We talk to
// it directly via REST rather than depending on the upstream Go SDK so the
// blast radius on the binary size + CVEs stays small.
package qdrant

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
)

// Client is safe for concurrent use.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Enabled reports whether a base URL is configured. Callers should short-circuit
// when false so the feature cleanly degrades.
func (c *Client) Enabled() bool {
	return c != nil && c.baseURL != ""
}

// Distance is the vector similarity metric Qdrant will use for a collection.
type Distance string

const (
	Cosine Distance = "Cosine"
	Dot    Distance = "Dot"
	Euclid Distance = "Euclid"
)

// CollectionExists returns true when the collection already exists; false
// otherwise. It translates 404 into false without surfacing an error.
func (c *Client) CollectionExists(ctx context.Context, name string) (bool, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/collections/"+name, nil)
	if err != nil {
		return false, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return true, nil
	}
	return false, decodeError(res)
}

// EnsureCollection creates the collection if missing. Size is the vector
// dimension; providers differ (OpenAI 3-small=1536, Google text-embedding-004=768).
// Changing size after creation is not supported — the caller must drop + recreate.
func (c *Client) EnsureCollection(ctx context.Context, name string, size int, distance Distance) error {
	exists, err := c.CollectionExists(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	body := map[string]any{
		"vectors": map[string]any{
			"size":     size,
			"distance": string(distance),
		},
	}
	raw, _ := json.Marshal(body)
	req, err := c.newRequest(ctx, http.MethodPut, "/collections/"+name, raw)
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
	return decodeError(res)
}

// DropCollection removes the whole collection (used when a user switches to
// an embedding model with a different vector size).
func (c *Client) DropCollection(ctx context.Context, name string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, "/collections/"+name, nil)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return decodeError(res)
}

type Point struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload,omitempty"`
}

// Upsert sends a batch of points (create-or-replace).
func (c *Client) Upsert(ctx context.Context, collection string, points []Point) error {
	body := map[string]any{"points": points}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := c.newRequest(ctx, http.MethodPut, "/collections/"+collection+"/points?wait=true", raw)
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
	return decodeError(res)
}

// DeleteByFileID removes all points whose payload.file_id matches. Used on
// hard-delete to avoid stale hits in semantic search.
func (c *Client) DeleteByFileID(ctx context.Context, collection, fileID string) error {
	body := map[string]any{
		"filter": map[string]any{
			"must": []map[string]any{
				{"key": "file_id", "match": map[string]any{"value": fileID}},
			},
		},
	}
	raw, _ := json.Marshal(body)
	req, err := c.newRequest(ctx, http.MethodPost, "/collections/"+collection+"/points/delete?wait=true", raw)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}
	return decodeError(res)
}

type SearchHit struct {
	ID      string         `json:"id"`
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload"`
}

// Search returns top-k nearest neighbours by vector. Optional filter matches
// payload fields; pass nil to search unfiltered.
func (c *Client) Search(ctx context.Context, collection string, vector []float32, limit int, filter map[string]any) ([]SearchHit, error) {
	if limit <= 0 {
		limit = 10
	}
	body := map[string]any{
		"vector":       vector,
		"limit":        limit,
		"with_payload": true,
	}
	if filter != nil {
		body["filter"] = filter
	}
	raw, _ := json.Marshal(body)
	req, err := c.newRequest(ctx, http.MethodPost, "/collections/"+collection+"/points/search", raw)
	if err != nil {
		return nil, err
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, decodeError(res)
	}
	var decoded struct {
		Result []SearchHit `json:"result"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	return decoded.Result, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body []byte) (*http.Request, error) {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	if len(body) > 0 {
		req.Header.Set("content-type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("api-key", c.apiKey)
	}
	return req, nil
}

func decodeError(res *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(res.Body, 8192))
	if len(body) == 0 {
		return fmt.Errorf("qdrant %d: %s", res.StatusCode, res.Status)
	}
	var env struct {
		Status struct {
			Error string `json:"error"`
		} `json:"status"`
	}
	if err := json.Unmarshal(body, &env); err == nil && env.Status.Error != "" {
		return fmt.Errorf("qdrant %d: %s", res.StatusCode, env.Status.Error)
	}
	return errors.New(strings.TrimSpace(string(body)))
}
