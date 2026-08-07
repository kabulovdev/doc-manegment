package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/http/dto"
	mw "gitlab.com/docuemnt_manegment/backend/internal/http/middleware"
	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
	"gitlab.com/docuemnt_manegment/backend/internal/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AIHandler struct {
	svc       *services.AIService
	features  *services.AIFeatures
	extractor *services.AIExtractorService
	validate  *validator.Validate
}

func NewAIHandler(svc *services.AIService, features *services.AIFeatures, extractor *services.AIExtractorService, v *validator.Validate) *AIHandler {
	return &AIHandler{svc: svc, features: features, extractor: extractor, validate: v}
}

func (h *AIHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	list, err := h.svc.List(r.Context(), uid)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.AIConfigResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.AIConfigToResponse(&list[i]))
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *AIHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	c, err := h.svc.Find(r.Context(), uid, id)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.AIConfigToResponse(c))
}

func (h *AIHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.CreateAIConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	caps, err := convertCapabilities(req.Capabilities)
	if err != nil {
		respond.Err(w, err)
		return
	}
	cfg, err := h.svc.Create(r.Context(), uid, services.CreateAIConfigInput{
		DisplayName:     req.DisplayName,
		Provider:        domain.AIProviderKind(req.Provider),
		BaseURL:         req.BaseURL,
		APIKey:          req.APIKey,
		ChatModel:       req.ChatModel,
		EmbedModel:      req.EmbedModel,
		TranscribeModel: req.TranscribeModel,
		Capabilities:    caps,
		DefaultChat:     req.DefaultChat,
		DefaultEmbed:    req.DefaultEmbed,
		DefaultTrans:    req.DefaultTrans,
	})
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusCreated, dto.AIConfigToResponse(cfg))
}

func (h *AIHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.UpdateAIConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	in := services.UpdateAIConfigInput{
		DisplayName:     req.DisplayName,
		ChatModel:       req.ChatModel,
		EmbedModel:      req.EmbedModel,
		TranscribeModel: req.TranscribeModel,
	}
	if req.Capabilities != nil {
		caps, err := convertCapabilities(*req.Capabilities)
		if err != nil {
			respond.Err(w, err)
			return
		}
		in.Capabilities = &caps
	}
	cfg, err := h.svc.Update(r.Context(), uid, id, in)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.AIConfigToResponse(cfg))
}

func (h *AIHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), uid, id); err != nil {
		respond.Err(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AIHandler) Test(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Test(r.Context(), uid, id); err != nil {
		respond.JSON(w, http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *AIHandler) SetDefault(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	id, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.SetDefaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.svc.SetDefault(r.Context(), uid, id, domain.AICapability(req.Capability)); err != nil {
		respond.Err(w, err)
		return
	}
	c, err := h.svc.Find(r.Context(), uid, id)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.AIConfigToResponse(c))
}

func (h *AIHandler) Usage(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	totals, err := h.svc.UsageTotals(r.Context(), uid)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := make([]dto.AIUsageTotalsResponse, 0, len(totals))
	for id, t := range totals {
		out = append(out, dto.AIUsageTotalsResponse{
			AIConfigID: id.Hex(),
			TokensIn:   t.TokensIn,
			TokensOut:  t.TokensOut,
			Calls:      t.Calls,
		})
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *AIHandler) SuggestFields(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	fields, meta, err := h.features.SuggestFields(r.Context(), uid, fileID)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := dto.SuggestFieldsResponse{
		Fields: make([]dto.SuggestedFieldDTO, 0, len(fields)),
		Meta:   aiMeta(meta),
	}
	for _, f := range fields {
		out.Fields = append(out.Fields, dto.SuggestedFieldDTO{
			Key: f.Key, Value: f.Value, Type: f.Type, Confidence: f.Confidence,
		})
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *AIHandler) SuggestTags(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	tags, meta, err := h.features.SuggestTags(r.Context(), uid, fileID)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := dto.SuggestTagsResponse{
		Tags: make([]dto.SuggestedTagDTO, 0, len(tags)),
		Meta: aiMeta(meta),
	}
	for _, t := range tags {
		out.Tags = append(out.Tags, dto.SuggestedTagDTO{Name: t.Name, Confidence: t.Confidence})
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *AIHandler) SuggestName(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	name, meta, err := h.features.SuggestName(r.Context(), uid, fileID)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.SuggestNameResponse{Name: name, Meta: aiMeta(meta)})
}

func (h *AIHandler) Search(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.SemanticSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	hits, err := h.features.SemanticSearch(r.Context(), uid, req.Query, limit)
	if err != nil {
		respond.Err(w, err)
		return
	}
	out := dto.SemanticSearchResponse{Hits: make([]dto.SemanticSearchHit, 0, len(hits))}
	for _, hit := range hits {
		out.Hits = append(out.Hits, dto.SemanticSearchHit{
			FileID: hit.FileID, Score: hit.Score, Name: hit.Name, Mime: hit.Mime,
			FolderID: hit.FolderID, TextPreview: hit.TextPreview,
		})
	}
	respond.JSON(w, http.StatusOK, out)
}

func (h *AIHandler) Ask(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	var req dto.AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	res, err := h.features.Ask(r.Context(), uid, req.Question, req.Limit)
	if err != nil {
		respond.Err(w, err)
		return
	}
	cites := make([]dto.AskCitationDTO, 0, len(res.Citations))
	for _, c := range res.Citations {
		cites = append(cites, dto.AskCitationDTO{
			Index: c.Index, FileID: c.FileID, Name: c.Name, Mime: c.Mime,
			Score: c.Score, Snippet: c.Snippet,
		})
	}
	respond.JSON(w, http.StatusOK, dto.AskResponse{
		Answer:    res.Answer,
		Citations: cites,
		Provider:  res.Provider,
		Model:     res.Model,
		TokensIn:  res.TokensIn,
		TokensOut: res.TokensOut,
		LatencyMS: res.LatencyMS,
	})
}

func (h *AIHandler) SuggestFolder(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	res, err := h.features.SuggestFolder(r.Context(), uid, fileID)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.SuggestFolderResponse{
		FolderID:   res.FolderID,
		FolderName: res.FolderName,
		Score:      res.Score,
	})
}

// ProcessDocument runs the one-shot vision-AI extraction for a file. The
// result is persisted on the File document so downstream tools can read the
// text cheaply.
func (h *AIHandler) ProcessDocument(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.ProcessDocumentRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respond.ValidationErr(w, err)
			return
		}
	}
	input := services.ProcessInput{}
	if req.ProviderID != "" {
		pid, err := primitive.ObjectIDFromHex(req.ProviderID)
		if err != nil {
			respond.Err(w, domain.ErrInvalidInput)
			return
		}
		input.ProviderID = &pid
	}
	ext, err := h.extractor.Process(r.Context(), uid, fileID, input)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.AIExtractionToResponse(ext))
}

// FileChat answers a single question grounded in one document's cached AI
// extraction text. No RAG — the document is small enough to fit in the
// chat call directly.
func (h *AIHandler) FileChat(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	var req dto.FileChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		respond.ValidationErr(w, err)
		return
	}
	res, err := h.features.ChatWithFile(r.Context(), uid, fileID, req.Question)
	if err != nil {
		respond.Err(w, err)
		return
	}
	cites := make([]dto.AskCitationDTO, 0, len(res.Citations))
	for _, c := range res.Citations {
		cites = append(cites, dto.AskCitationDTO{
			Index: c.Index, FileID: c.FileID, Name: c.Name, Mime: c.Mime,
			Score: c.Score, Snippet: c.Snippet,
		})
	}
	respond.JSON(w, http.StatusOK, dto.AskResponse{
		Answer:    res.Answer,
		Citations: cites,
		Provider:  res.Provider,
		Model:     res.Model,
		TokensIn:  res.TokensIn,
		TokensOut: res.TokensOut,
		LatencyMS: res.LatencyMS,
	})
}

// GetExtraction returns the cached extraction state for a file — used by the
// AI panel to decide whether to offer "Process" or the tool buttons.
func (h *AIHandler) GetExtraction(w http.ResponseWriter, r *http.Request) {
	uid, _ := mw.UserIDFromContext(r.Context())
	fileID, ok := objectID(w, r)
	if !ok {
		return
	}
	ext, err := h.extractor.Get(r.Context(), uid, fileID)
	if err != nil {
		respond.Err(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, dto.AIExtractionToResponse(ext))
}

func aiMeta(m services.SuggestionsResult) dto.AISuggestionsMeta {
	return dto.AISuggestionsMeta{
		Provider:         m.Provider,
		Model:            m.Model,
		TokensIn:         m.TokensIn,
		TokensOut:        m.TokensOut,
		LatencyMS:        m.LatencyMS,
		ExtractionStatus: string(m.ExtractionStatus),
	}
}

func convertCapabilities(in []string) ([]domain.AICapability, error) {
	out := make([]domain.AICapability, 0, len(in))
	for _, raw := range in {
		c := domain.AICapability(raw)
		if !c.Valid() {
			return nil, domain.ErrInvalidInput
		}
		out = append(out, c)
	}
	return out, nil
}
