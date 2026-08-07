package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/adapters/ai"
	"gitlab.com/docuemnt_manegment/backend/internal/adapters/storage/ssrf"
	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AIService owns the CRUD + resolution logic for per-user AI provider configs.
// It mirrors StorageService: encrypted creds on disk, SSRF-guarded HTTP
// transport for user-supplied base URLs, test-before-save, and a thin cache
// of decrypted provider instances keyed by config id.
type AIService struct {
	repo       ports.AIConfigRepo
	usage      ports.AIUsageRepo
	enc        ports.Encryptor
	ssrfPolicy ssrf.Policy
	httpClient *http.Client

	cacheMu sync.Mutex
	cache   map[primitive.ObjectID]ports.AIProvider
}

func NewAIService(repo ports.AIConfigRepo, usage ports.AIUsageRepo, enc ports.Encryptor, policy ssrf.Policy) *AIService {
	transport := ssrf.GuardedTransport(policy)
	return &AIService{
		repo:       repo,
		usage:      usage,
		enc:        enc,
		ssrfPolicy: policy,
		httpClient: &http.Client{Transport: transport, Timeout: policy.RequestTimeout},
		cache:      map[primitive.ObjectID]ports.AIProvider{},
	}
}

type CreateAIConfigInput struct {
	DisplayName     string
	Provider        domain.AIProviderKind
	BaseURL         string
	APIKey          string
	ChatModel       string
	EmbedModel      string
	TranscribeModel string
	Capabilities    []domain.AICapability
	DefaultChat     bool
	DefaultEmbed    bool
	DefaultTrans    bool
}

type UpdateAIConfigInput struct {
	DisplayName     *string
	ChatModel       *string
	EmbedModel      *string
	TranscribeModel *string
	Capabilities    *[]domain.AICapability
}

func (s *AIService) Create(ctx context.Context, userID primitive.ObjectID, in CreateAIConfigInput) (*domain.AIConfig, error) {
	name := strings.TrimSpace(in.DisplayName)
	if name == "" || len(name) > 100 {
		return nil, fmt.Errorf("%w: display_name", domain.ErrInvalidInput)
	}
	if !in.Provider.Valid() {
		return nil, fmt.Errorf("%w: provider", domain.ErrInvalidInput)
	}
	if strings.TrimSpace(in.APIKey) == "" && in.Provider != domain.AIProviderOllama {
		return nil, fmt.Errorf("%w: api_key required", domain.ErrInvalidInput)
	}
	if err := s.validateBaseURL(in.Provider, in.BaseURL); err != nil {
		return nil, err
	}
	caps, err := normalizeCapabilities(in.Capabilities, in.Provider)
	if err != nil {
		return nil, err
	}

	ct, nonce, err := s.enc.Encrypt([]byte(in.APIKey))
	if err != nil {
		return nil, err
	}

	cfg := &domain.AIConfig{
		UserID:          userID,
		DisplayName:     name,
		Provider:        in.Provider,
		BaseURL:         strings.TrimRight(strings.TrimSpace(in.BaseURL), "/"),
		APIKeyCT:        ct,
		APIKeyNonce:     nonce,
		KeyVersion:      s.enc.KeyVersion(),
		KeyPrefix:       keyPrefix(in.APIKey),
		ChatModel:       strings.TrimSpace(in.ChatModel),
		EmbedModel:      strings.TrimSpace(in.EmbedModel),
		TranscribeModel: strings.TrimSpace(in.TranscribeModel),
		Capabilities:    caps,
	}

	// Ping before persisting so users get immediate feedback on bad keys.
	provider, err := ai.NewProvider(cfg, in.APIKey, s.httpClient)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := provider.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("%w: provider test failed: %v", domain.ErrInvalidInput, err)
	}

	now := time.Now().UTC()
	cfg.LastTestedAt = &now

	if err := s.repo.Create(ctx, cfg); err != nil {
		return nil, err
	}

	// Apply default flags after persisting (requires id to exist).
	if in.DefaultChat {
		_ = s.SetDefault(ctx, userID, cfg.ID, domain.AICapChat)
	}
	if in.DefaultEmbed {
		_ = s.SetDefault(ctx, userID, cfg.ID, domain.AICapEmbed)
	}
	if in.DefaultTrans {
		_ = s.SetDefault(ctx, userID, cfg.ID, domain.AICapTranscribe)
	}

	return s.repo.Find(ctx, userID, cfg.ID)
}

func (s *AIService) List(ctx context.Context, userID primitive.ObjectID) ([]domain.AIConfig, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *AIService) Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.AIConfig, error) {
	return s.repo.Find(ctx, userID, id)
}

func (s *AIService) Update(ctx context.Context, userID, id primitive.ObjectID, in UpdateAIConfigInput) (*domain.AIConfig, error) {
	set := bson.M{}
	if in.DisplayName != nil {
		name := strings.TrimSpace(*in.DisplayName)
		if name == "" || len(name) > 100 {
			return nil, fmt.Errorf("%w: display_name", domain.ErrInvalidInput)
		}
		set["display_name"] = name
	}
	if in.ChatModel != nil {
		set["chat_model"] = strings.TrimSpace(*in.ChatModel)
	}
	if in.EmbedModel != nil {
		set["embed_model"] = strings.TrimSpace(*in.EmbedModel)
	}
	if in.TranscribeModel != nil {
		set["transcribe_model"] = strings.TrimSpace(*in.TranscribeModel)
	}
	if in.Capabilities != nil {
		caps, err := normalizeCapabilities(*in.Capabilities, "")
		if err != nil {
			return nil, err
		}
		set["capabilities"] = caps
	}
	if len(set) == 0 {
		return s.repo.Find(ctx, userID, id)
	}
	if err := s.repo.Update(ctx, userID, id, set); err != nil {
		return nil, err
	}
	s.invalidate(id)
	return s.repo.Find(ctx, userID, id)
}

func (s *AIService) Delete(ctx context.Context, userID, id primitive.ObjectID) error {
	s.invalidate(id)
	return s.repo.Delete(ctx, userID, id)
}

func (s *AIService) SetDefault(ctx context.Context, userID, id primitive.ObjectID, capability domain.AICapability) error {
	cfg, err := s.repo.Find(ctx, userID, id)
	if err != nil {
		return err
	}
	if !hasCapability(cfg.Capabilities, capability) {
		return fmt.Errorf("%w: config does not declare %s", domain.ErrInvalidInput, capability)
	}
	if err := s.repo.ClearDefault(ctx, userID, capability); err != nil {
		return err
	}
	field, ok := defaultField(capability)
	if !ok {
		return domain.ErrInvalidInput
	}
	return s.repo.Update(ctx, userID, id, bson.M{field: true})
}

// Test re-runs Ping and records LastError / LastTestedAt on the config.
func (s *AIService) Test(ctx context.Context, userID, id primitive.ObjectID) error {
	cfg, err := s.repo.Find(ctx, userID, id)
	if err != nil {
		return err
	}
	key, err := s.decrypt(cfg)
	if err != nil {
		return err
	}
	provider, err := ai.NewProvider(cfg, key, s.httpClient)
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	now := time.Now().UTC()
	if err := provider.Ping(pingCtx); err != nil {
		_ = s.repo.Update(ctx, userID, id, bson.M{
			"last_error":     err.Error(),
			"last_tested_at": now,
		})
		return err
	}
	_ = s.repo.Update(ctx, userID, id, bson.M{
		"last_error":     "",
		"last_tested_at": now,
	})
	return nil
}

// ResolveProvider picks the provider for a given capability. Explicit cfgID
// wins; otherwise the user's marked default; otherwise the first config that
// declares the capability. Returns ErrNoAIConfig when nothing matches.
func (s *AIService) ResolveProvider(ctx context.Context, userID primitive.ObjectID, capability domain.AICapability, cfgID *primitive.ObjectID) (ports.AIProvider, *domain.AIConfig, error) {
	list, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	var chosen *domain.AIConfig
	if cfgID != nil && !cfgID.IsZero() {
		for i := range list {
			if list[i].ID == *cfgID && hasCapability(list[i].Capabilities, capability) {
				chosen = &list[i]
				break
			}
		}
	}
	if chosen == nil {
		for i := range list {
			if isDefaultFor(&list[i], capability) && hasCapability(list[i].Capabilities, capability) {
				chosen = &list[i]
				break
			}
		}
	}
	if chosen == nil {
		for i := range list {
			if hasCapability(list[i].Capabilities, capability) {
				chosen = &list[i]
				break
			}
		}
	}
	if chosen == nil {
		return nil, nil, domain.ErrNoAIConfig
	}

	s.cacheMu.Lock()
	if p, ok := s.cache[chosen.ID]; ok {
		s.cacheMu.Unlock()
		return p, chosen, nil
	}
	s.cacheMu.Unlock()

	key, err := s.decrypt(chosen)
	if err != nil {
		return nil, nil, err
	}
	provider, err := ai.NewProvider(chosen, key, s.httpClient)
	if err != nil {
		return nil, nil, err
	}
	s.cacheMu.Lock()
	s.cache[chosen.ID] = provider
	s.cacheMu.Unlock()
	return provider, chosen, nil
}

func (s *AIService) UsageTotals(ctx context.Context, userID primitive.ObjectID) (map[primitive.ObjectID]ports.AIUsageTotals, error) {
	return s.usage.TotalsByConfig(ctx, userID)
}

// RecordUsage persists a usage row and bumps the per-config aggregate. Safe
// to call from background workers; never returns an error that should halt
// the caller.
func (s *AIService) RecordUsage(ctx context.Context, userID, cfgID primitive.ObjectID, feature, model string, tokensIn, tokensOut int, status int, latency time.Duration) {
	now := time.Now().UTC()
	_ = s.usage.Create(ctx, &domain.AIUsage{
		UserID:     userID,
		AIConfigID: cfgID,
		Feature:    feature,
		Model:      model,
		TokensIn:   tokensIn,
		TokensOut:  tokensOut,
		Status:     status,
		LatencyMS:  latency.Milliseconds(),
		CreatedAt:  now,
	})
	_ = s.repo.IncrementUsage(ctx, cfgID, int64(tokensIn), int64(tokensOut), now)
}

func (s *AIService) invalidate(id primitive.ObjectID) {
	s.cacheMu.Lock()
	delete(s.cache, id)
	s.cacheMu.Unlock()
}

func (s *AIService) decrypt(cfg *domain.AIConfig) (string, error) {
	if len(cfg.APIKeyCT) == 0 {
		return "", nil
	}
	pt, err := s.enc.Decrypt(cfg.APIKeyCT, cfg.APIKeyNonce)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func (s *AIService) validateBaseURL(kind domain.AIProviderKind, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		// SaaS providers use a fixed baseURL baked into the adapter; an empty
		// value is fine.
		return nil
	}
	if kind != domain.AIProviderOpenAICompat && kind != domain.AIProviderOllama {
		return fmt.Errorf("%w: base_url is only valid for openai-compatible or ollama", domain.ErrInvalidInput)
	}
	if _, err := ssrf.ValidateEndpoint(raw, s.ssrfPolicy); err != nil {
		return fmt.Errorf("%w: base_url: %v", domain.ErrInvalidInput, err)
	}
	return nil
}

func hasCapability(have []domain.AICapability, want domain.AICapability) bool {
	for _, c := range have {
		if c == want {
			return true
		}
	}
	return false
}

func isDefaultFor(c *domain.AIConfig, cap domain.AICapability) bool {
	switch cap {
	case domain.AICapChat:
		return c.IsDefaultChat
	case domain.AICapEmbed:
		return c.IsDefaultEmbed
	case domain.AICapTranscribe:
		return c.IsDefaultTranscribe
	}
	return false
}

func defaultField(cap domain.AICapability) (string, bool) {
	switch cap {
	case domain.AICapChat:
		return "is_default_chat", true
	case domain.AICapEmbed:
		return "is_default_embed", true
	case domain.AICapTranscribe:
		return "is_default_transcribe", true
	}
	return "", false
}

func normalizeCapabilities(in []domain.AICapability, fallback domain.AIProviderKind) ([]domain.AICapability, error) {
	if len(in) == 0 {
		// Sensible defaults per provider so the UI doesn't have to set them
		// explicitly for the common case.
		switch fallback {
		case domain.AIProviderAnthropic:
			return []domain.AICapability{domain.AICapChat}, nil
		case domain.AIProviderOpenAI:
			return []domain.AICapability{domain.AICapChat, domain.AICapEmbed, domain.AICapTranscribe}, nil
		case domain.AIProviderGoogle:
			return []domain.AICapability{domain.AICapChat, domain.AICapEmbed}, nil
		case domain.AIProviderOpenAICompat:
			return []domain.AICapability{domain.AICapChat}, nil
		case domain.AIProviderOllama:
			return []domain.AICapability{domain.AICapChat, domain.AICapEmbed}, nil
		case "":
			return []domain.AICapability{}, nil
		}
	}
	seen := map[domain.AICapability]struct{}{}
	out := make([]domain.AICapability, 0, len(in))
	for _, c := range in {
		if !c.Valid() {
			return nil, fmt.Errorf("%w: unknown capability %q", domain.ErrInvalidInput, c)
		}
		if _, dup := seen[c]; dup {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out, nil
}

func keyPrefix(key string) string {
	k := strings.TrimSpace(key)
	if len(k) <= 6 {
		return k
	}
	return k[:6]
}
