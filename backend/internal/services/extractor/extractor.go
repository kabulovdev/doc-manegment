// Package extractor produces plain-text representations of file contents for
// downstream AI features. A separate service from FileService because it
// (a) caches per-file, (b) has its own error modes, (c) may shell out to
// external binaries (pdftotext).
package extractor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/domain"
	"gitlab.com/docuemnt_manegment/backend/internal/ports"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileSource is the minimal surface extractor needs from FileService, kept as
// an interface here to avoid a services→services import cycle when feature
// services (e.g. ai_features) depend on extractor.
type FileSource interface {
	Find(ctx context.Context, userID, id primitive.ObjectID) (*domain.File, error)
	Stream(ctx context.Context, userID, id primitive.ObjectID, rng *ports.Range) (io.ReadCloser, *domain.File, ports.ObjectInfo, error)
}

// maxBytes caps how much of a file we'll read into memory. 10 MB is generous
// for any single-pass LLM context; larger files should use a chunked pipeline
// that isn't in Wave 1.
const maxBytes = 10 * 1024 * 1024

// maxTextChars caps how much extracted text we persist. Longer text is
// truncated; we keep what fits into a typical chat context window.
const maxTextChars = 40000

type Service struct {
	repo  ports.ExtractionRepo
	files FileSource
}

func NewService(repo ports.ExtractionRepo, files FileSource) *Service {
	return &Service{repo: repo, files: files}
}

// ExtractResult is returned synchronously to callers; cache hits skip the
// download + engine step.
type ExtractResult struct {
	Status domain.ExtractionStatus
	Text   string
	Mime   string
	Engine string
	Error  string
}

// Extract returns the text for a file, using cached output when the file
// hasn't been re-uploaded. On cache miss it downloads the file body, routes
// by mime to the right extraction engine, and stores the result.
func (s *Service) Extract(ctx context.Context, userID, fileID primitive.ObjectID) (*ExtractResult, error) {
	file, err := s.files.Find(ctx, userID, fileID)
	if err != nil {
		return nil, err
	}
	cached, _ := s.repo.FindByFile(ctx, userID, fileID)
	if cached != nil && sameUpload(cached.UploadedAt, file.UploadedAt) && cached.Status != domain.ExtractionPending {
		return &ExtractResult{
			Status: cached.Status,
			Text:   cached.Text,
			Mime:   cached.Mime,
			Engine: cached.Engine,
			Error:  cached.Error,
		}, nil
	}

	start := time.Now()
	result, engine, extractErr := s.runExtraction(ctx, userID, fileID, file.MimeType)

	entry := &domain.FileExtraction{
		UserID:     userID,
		FileID:     fileID,
		UploadedAt: file.UploadedAt,
		Mime:       file.MimeType,
		Engine:     engine,
		DurationMS: time.Since(start).Milliseconds(),
	}
	if extractErr != nil {
		if errors.Is(extractErr, errUnsupported) {
			entry.Status = domain.ExtractionUnsupported
			entry.Error = extractErr.Error()
		} else {
			entry.Status = domain.ExtractionError
			entry.Error = extractErr.Error()
		}
	} else {
		text := truncate(result, maxTextChars)
		entry.Status = domain.ExtractionReady
		entry.Text = text
		entry.CharCount = len([]rune(text))
	}
	_ = s.repo.Upsert(ctx, entry)

	return &ExtractResult{
		Status: entry.Status,
		Text:   entry.Text,
		Mime:   entry.Mime,
		Engine: entry.Engine,
		Error:  entry.Error,
	}, nil
}

var errUnsupported = errors.New("extraction not supported for this mime type")

func (s *Service) runExtraction(ctx context.Context, userID, fileID primitive.ObjectID, mime string) (string, string, error) {
	mime = strings.ToLower(mime)
	switch {
	case strings.HasPrefix(mime, "text/"),
		mime == "application/json",
		mime == "application/xml",
		mime == "application/x-yaml":
		buf, err := s.downloadCapped(ctx, userID, fileID)
		if err != nil {
			return "", "plain-reader", err
		}
		return string(buf), "plain-reader", nil
	case mime == "application/pdf":
		buf, err := s.downloadCapped(ctx, userID, fileID)
		if err != nil {
			return "", "pdftotext", err
		}
		text, err := runPDFToText(ctx, buf)
		if err != nil {
			return "", "pdftotext", err
		}
		return text, "pdftotext", nil
	}
	return "", "", errUnsupported
}

func (s *Service) downloadCapped(ctx context.Context, userID, fileID primitive.ObjectID) ([]byte, error) {
	rc, _, _, err := s.files.Stream(ctx, userID, fileID, nil)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(io.LimitReader(rc, maxBytes+1))
}

// runPDFToText shells out to `pdftotext -layout -` reading from stdin. When
// the binary isn't installed we return an explicit error — the UI degrades
// to "PDF extraction unavailable" rather than hanging.
func runPDFToText(ctx context.Context, body []byte) (string, error) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", errors.New("pdftotext binary not installed; install poppler-utils to enable PDF extraction")
	}
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, "pdftotext", "-layout", "-enc", "UTF-8", "-", "-")
	cmd.Stdin = bytes.NewReader(body)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if cmdCtx.Err() != nil {
			return "", fmt.Errorf("pdftotext timed out: %w", cmdCtx.Err())
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("pdftotext failed: %s", msg)
	}
	return out.String(), nil
}

func sameUpload(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

func truncate(s string, max int) string {
	if max <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// suppress the unused-import warning when building tests that import os.
var _ = os.TempDir
