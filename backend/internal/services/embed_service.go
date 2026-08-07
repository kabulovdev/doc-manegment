package services

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/adapters/qdrant"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/services/extractor"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmbedService owns the "document → vector → Qdrant" pipeline. It uses the
// user's own embed-capable AIConfig (BYOK) and stores vectors in a per-user
// Qdrant collection so the dimensionality can differ between users.
type EmbedService struct {
	ai        *AIService
	extractor *extractor.Service
	qdrant    *qdrant.Client
	files     *FileService

	dimMu   sync.Mutex
	dimByID map[primitive.ObjectID]int
}

func NewEmbedService(ai *AIService, extractor *extractor.Service, qdr *qdrant.Client, files *FileService) *EmbedService {
	return &EmbedService{
		ai:        ai,
		extractor: extractor,
		qdrant:    qdr,
		files:     files,
		dimByID:   map[primitive.ObjectID]int{},
	}
}

// Enabled reports whether Qdrant is configured and we can do vector work.
// Features that depend on it should short-circuit when false.
func (s *EmbedService) Enabled() bool {
	return s != nil && s.qdrant != nil && s.qdrant.Enabled()
}

// CollectionName returns the per-user collection. Keeping users isolated lets
// each pick their own embed model without cross-collision on vector size.
func (s *EmbedService) CollectionName(userID primitive.ObjectID) string {
	return "files_" + userID.Hex()
}

// IndexFile chunks the file's text and upserts one point per chunk into
// Qdrant. Priority order for the source text:
//  1. File.AIExtraction.Text (written by the vision-AI extractor — clean,
//     structured markdown that works even for scanned PDFs).
//  2. extractor.Extract — the raw pdftotext / plain-text fallback.
//
// Existing chunks for the file are deleted first to keep the collection
// consistent across re-indexes. Safe to call concurrently; idempotent.
// Returns the vector dimension on success.
func (s *EmbedService) IndexFile(ctx context.Context, userID, fileID primitive.ObjectID) (int, error) {
	if !s.Enabled() {
		return 0, errors.New("qdrant not configured")
	}
	file, err := s.files.Find(ctx, userID, fileID)
	if err != nil {
		return 0, err
	}

	var source string
	if file.AIExtraction.Status == domain.AIExtractionReady && file.AIExtraction.Text != "" {
		source = file.AIExtraction.Text
	} else {
		ext, err := s.extractor.Extract(ctx, userID, fileID)
		if err != nil {
			return 0, err
		}
		if ext.Status != domain.ExtractionReady || ext.Text == "" {
			return 0, fmt.Errorf("no text to embed (status=%s)", ext.Status)
		}
		source = ext.Text
	}

	chunks := ChunkMarkdown(source)
	if len(chunks) == 0 {
		return 0, errors.New("no chunks produced from document text")
	}

	provider, cfg, err := s.ai.ResolveProvider(ctx, userID, domain.AICapEmbed, nil)
	if err != nil {
		return 0, err
	}

	texts := make([]string, 0, len(chunks))
	for _, ch := range chunks {
		texts = append(texts, buildEmbedText(file.Name, ch.Text))
	}

	start := time.Now()
	vectors, tokensIn, err := provider.Embed(ctx, texts)
	s.ai.RecordUsage(ctx, userID, cfg.ID, "embed_index", cfg.EmbedModel, tokensIn, 0, 200, time.Since(start))
	if err != nil {
		return 0, err
	}
	if len(vectors) != len(chunks) || len(vectors[0]) == 0 {
		return 0, errors.New("provider returned a mismatched or empty embedding batch")
	}

	dim := len(vectors[0])
	collection := s.CollectionName(userID)
	if err := s.qdrant.EnsureCollection(ctx, collection, dim, qdrant.Cosine); err != nil {
		return 0, err
	}
	// Drop old points for this file so re-indexing doesn't leave orphans
	// behind (e.g. when a re-extraction yields fewer chunks than before).
	if err := s.qdrant.DeleteByFileID(ctx, collection, file.ID.Hex()); err != nil {
		// Non-fatal — a new upsert still overwrites any ID collision.
		slog.Warn("embed delete-before-upsert failed", "err", err.Error())
	}

	points := make([]qdrant.Point, 0, len(chunks))
	for i, ch := range chunks {
		payload := map[string]any{
			"file_id":       file.ID.Hex(),
			"chunk_index":   ch.Index,
			"chunk_text":    ch.Text,
			"section_title": ch.SectionTitle,
			"name":          file.Name,
			"mime":          file.MimeType,
			"size_bytes":    file.SizeBytes,
			"embed_model":   cfg.EmbedModel,
			"embed_config":  cfg.ID.Hex(),
			"text_preview":  preview(ch.Text),
		}
		if file.FolderID != nil {
			payload["folder_id"] = file.FolderID.Hex()
		}
		points = append(points, qdrant.Point{
			ID:      chunkPointID(file.ID.Hex(), ch.Index),
			Vector:  vectors[i],
			Payload: payload,
		})
	}
	if err := s.qdrant.Upsert(ctx, collection, points); err != nil {
		return 0, err
	}
	s.cacheDim(cfg.ID, dim)
	return dim, nil
}

// chunkPointID generates a deterministic UUID-formatted ID that Qdrant will
// accept. We MD5 the `fileID:chunkIndex` pair and format as 8-4-4-4-12 so
// re-indexing the same chunk lands on the same point (upsert stays idempotent).
func chunkPointID(fileIDHex string, index int) string {
	h := md5.Sum([]byte(fileIDHex + ":" + strconv.Itoa(index)))
	s := hex.EncodeToString(h[:])
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:32]
}

// IndexFileAsync runs IndexFile in the background and logs the outcome. Use
// this from upload handlers so the user's request returns immediately.
func (s *EmbedService) IndexFileAsync(userID, fileID primitive.ObjectID) {
	if !s.Enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if _, err := s.IndexFile(ctx, userID, fileID); err != nil {
			slog.Info("embed index skipped",
				"user_id", userID.Hex(),
				"file_id", fileID.Hex(),
				"err", err.Error(),
			)
		}
	}()
}

// RemoveFile drops a file's vector. Best-effort: we don't want bucket/mongo
// delete to fail just because Qdrant was offline for a moment.
func (s *EmbedService) RemoveFile(ctx context.Context, userID, fileID primitive.ObjectID) {
	if !s.Enabled() {
		return
	}
	collection := s.CollectionName(userID)
	if err := s.qdrant.DeleteByFileID(ctx, collection, fileID.Hex()); err != nil {
		slog.Info("embed delete failed", "user_id", userID.Hex(), "file_id", fileID.Hex(), "err", err.Error())
	}
}

// SearchResult is a single semantic-search hit enriched with payload metadata.
// ChunkText is populated for RAG-friendly retrieval (per-chunk text); older
// file-level points may only have TextPreview.
type SearchResult struct {
	FileID       string
	Score        float32
	Name         string
	Mime         string
	FolderID     string
	TextPreview  string
	ChunkIndex   int
	ChunkText    string
	SectionTitle string
}

// Search runs the query string through the user's embed provider and looks
// up nearest neighbours in their collection.
func (s *EmbedService) Search(ctx context.Context, userID primitive.ObjectID, query string, limit int) ([]SearchResult, error) {
	if !s.Enabled() {
		return nil, errors.New("qdrant not configured")
	}
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}
	provider, cfg, err := s.ai.ResolveProvider(ctx, userID, domain.AICapEmbed, nil)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	vectors, tokensIn, err := provider.Embed(ctx, []string{query})
	s.ai.RecordUsage(ctx, userID, cfg.ID, "embed_query", cfg.EmbedModel, tokensIn, 0, 200, time.Since(start))
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return nil, errors.New("provider returned no embedding")
	}
	hits, err := s.qdrant.Search(ctx, s.CollectionName(userID), vectors[0], limit, nil)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(hits))
	for _, h := range hits {
		out = append(out, SearchResult{
			FileID:       stringPayload(h.Payload, "file_id"),
			Score:        h.Score,
			Name:         stringPayload(h.Payload, "name"),
			Mime:         stringPayload(h.Payload, "mime"),
			FolderID:     stringPayload(h.Payload, "folder_id"),
			TextPreview:  stringPayload(h.Payload, "text_preview"),
			ChunkIndex:   intPayload(h.Payload, "chunk_index"),
			ChunkText:    stringPayload(h.Payload, "chunk_text"),
			SectionTitle: stringPayload(h.Payload, "section_title"),
		})
	}
	return out, nil
}

func intPayload(p map[string]any, key string) int {
	switch v := p[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

func (s *EmbedService) cacheDim(cfgID primitive.ObjectID, dim int) {
	s.dimMu.Lock()
	s.dimByID[cfgID] = dim
	s.dimMu.Unlock()
}

func stringPayload(p map[string]any, key string) string {
	if v, ok := p[key].(string); ok {
		return v
	}
	return ""
}

// buildEmbedText prefixes the filename so retrieval picks up keyword matches
// even when the body is short or mostly boilerplate.
func buildEmbedText(name, body string) string {
	body = strings.TrimSpace(body)
	if len(body) > 20000 {
		// Keep the head — the most information-dense part for most documents.
		body = body[:20000]
	}
	return name + "\n\n" + body
}

func preview(body string) string {
	body = strings.TrimSpace(body)
	if len(body) > 500 {
		body = body[:500]
	}
	return body
}
