// Package services — chunker.go splits a document's text into retrieval-
// friendly pieces for the RAG pipeline. We target ~800 tokens per chunk
// with a 100-token overlap, snapping to markdown section boundaries first
// (## headings from the vision extractor), then blank-line paragraphs.
package services

import (
	"regexp"
	"strings"
)

// Chunk is a single piece of a document with enough metadata to carry back
// a citation after retrieval.
type Chunk struct {
	Index        int
	Text         string
	SectionTitle string
}

// Chunking knobs. Tokens aren't counted exactly — we approximate at 4
// characters/token which matches typical English-ish prose closely enough
// for retrieval purposes. If a user embeds large Chinese or code-heavy
// corpora we'd revisit, but for document-management this is fine.
const (
	targetCharsPerChunk = 3200 // ~800 tokens
	minCharsPerChunk    = 400  // don't create tiny trailing chunks
	overlapChars        = 400  // ~100 tokens
)

var mdH2Re = regexp.MustCompile(`(?m)^##\s+.*$`)

// ChunkMarkdown splits markdown produced by the AI extractor into sections
// first, then falls back to blank-line paragraph splits when a single
// section is too large. Empty input returns nil.
func ChunkMarkdown(text string) []Chunk {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	sections := splitByH2(text)
	out := make([]Chunk, 0, len(sections))
	for _, sec := range sections {
		if len(sec.body) <= targetCharsPerChunk {
			out = append(out, Chunk{Text: sec.body, SectionTitle: sec.title})
			continue
		}
		for _, piece := range splitBigSection(sec.body) {
			out = append(out, Chunk{Text: piece, SectionTitle: sec.title})
		}
	}

	// Merge trailing tiny chunks into the previous one so we don't waste
	// retrieval slots on 50-char scraps.
	out = mergeTinyTrailers(out)

	for i := range out {
		out[i].Index = i
	}
	return out
}

type section struct {
	title string
	body  string
}

func splitByH2(text string) []section {
	locs := mdH2Re.FindAllStringIndex(text, -1)
	if len(locs) == 0 {
		return []section{{body: strings.TrimSpace(text)}}
	}
	out := make([]section, 0, len(locs)+1)
	// Preamble before the first heading (rare but possible).
	if locs[0][0] > 0 {
		pre := strings.TrimSpace(text[:locs[0][0]])
		if pre != "" {
			out = append(out, section{body: pre})
		}
	}
	for i, loc := range locs {
		headEnd := loc[1]
		title := strings.TrimSpace(strings.TrimPrefix(text[loc[0]:headEnd], "##"))
		nextStart := len(text)
		if i+1 < len(locs) {
			nextStart = locs[i+1][0]
		}
		body := strings.TrimSpace(text[loc[0]:nextStart])
		out = append(out, section{title: title, body: body})
	}
	return out
}

// splitBigSection breaks a single section into overlapping chunks by
// paragraph, then by sentence-ish boundaries if paragraphs themselves are
// too long. The overlap preserves context across chunk boundaries so a
// question landing near the split still retrieves both sides.
func splitBigSection(body string) []string {
	paragraphs := strings.Split(body, "\n\n")
	out := make([]string, 0)
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		out = append(out, strings.TrimSpace(cur.String()))
		cur.Reset()
	}
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if cur.Len()+len(p)+2 <= targetCharsPerChunk {
			if cur.Len() > 0 {
				cur.WriteString("\n\n")
			}
			cur.WriteString(p)
			continue
		}
		if cur.Len() > 0 {
			flush()
		}
		if len(p) <= targetCharsPerChunk {
			cur.WriteString(p)
			continue
		}
		// Paragraph itself is oversized — slice by char with overlap.
		for start := 0; start < len(p); {
			end := start + targetCharsPerChunk
			if end > len(p) {
				end = len(p)
			}
			out = append(out, strings.TrimSpace(p[start:end]))
			if end == len(p) {
				break
			}
			start = end - overlapChars
			if start < 0 {
				start = 0
			}
		}
	}
	flush()

	// Add overlap between consecutive chunks for smoother retrieval.
	if len(out) > 1 {
		withOverlap := make([]string, len(out))
		withOverlap[0] = out[0]
		for i := 1; i < len(out); i++ {
			prev := out[i-1]
			tail := prev
			if len(prev) > overlapChars {
				tail = prev[len(prev)-overlapChars:]
			}
			withOverlap[i] = strings.TrimSpace(tail) + "\n\n" + out[i]
		}
		return withOverlap
	}
	return out
}

func mergeTinyTrailers(chunks []Chunk) []Chunk {
	if len(chunks) < 2 {
		return chunks
	}
	last := chunks[len(chunks)-1]
	if len(last.Text) >= minCharsPerChunk {
		return chunks
	}
	prev := chunks[len(chunks)-2]
	prev.Text = strings.TrimSpace(prev.Text) + "\n\n" + strings.TrimSpace(last.Text)
	chunks[len(chunks)-2] = prev
	return chunks[:len(chunks)-1]
}
