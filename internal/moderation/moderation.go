// Package moderation runs the asynchronous AI compliance check that gates
// public gists. A gist submitted as public is stored privately with a pending
// status; a worker asks an OpenAI-compatible chat endpoint whether it may be
// published, then flips it to public or keeps it private with a stored reason.
package moderation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LiuShen-Fork/Qingyu-Gist/internal/db"
	"github.com/LiuShen-Fork/Qingyu-Gist/internal/settings"
	"github.com/rs/zerolog/log"
)

const (
	// maxDocumentBytes bounds the text sent to the provider so a single large
	// gist cannot blow up the request or the bill.
	maxDocumentBytes = 16 * 1024
	queueSize        = 128
)

var queue chan uint

// Start launches the moderation worker. It is safe to call when moderation is
// disabled: the worker only acts on queued gists and the setting is checked
// when a gist is enqueued.
func Start() {
	if queue != nil {
		return
	}
	queue = make(chan uint, queueSize)
	go worker()

	// Resume checks that were interrupted by a restart.
	if settings.ModerationEnabled() {
		if ids, err := db.GetPendingModerationGistIDs(); err == nil {
			for _, id := range ids {
				Enqueue(id)
			}
		}
	}
}

// Gate holds back a gist that was submitted as public until the worker has
// reviewed it. It returns true when the gist now needs to be enqueued; callers
// should still save the gist, then call Enqueue with its ID.
func Gate(gist *db.Gist) bool {
	if !settings.ModerationEnabled() || gist.Private != db.PublicVisibility || gist.ModerationStatus == db.ModerationApproved {
		return false
	}
	gist.Private = db.PrivateVisibility
	gist.ModerationStatus = db.ModerationPending
	return true
}

// Enqueue schedules a gist for moderation without blocking the request.
func Enqueue(gistID uint) {
	if queue == nil || !settings.ModerationEnabled() {
		return
	}
	select {
	case queue <- gistID:
	default:
		// The queue is saturated; run this one on its own goroutine rather than
		// leaving the gist pending forever.
		go moderate(gistID)
	}
}

func worker() {
	for id := range queue {
		moderate(id)
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type verdict struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

func moderate(gistID uint) {
	gist, err := db.GetGistByID(strconv.FormatUint(uint64(gistID), 10))
	if err != nil {
		log.Error().Err(err).Msgf("Moderation: cannot load gist %d", gistID)
		return
	}
	if gist.ModerationStatus != db.ModerationPending {
		return
	}

	allowed, reason, err := review(gist)
	if err != nil {
		log.Warn().Err(err).Msgf("Moderation: review failed for gist %d", gistID)
		if settings.ModerationFailClosed() {
			finish(gist, db.ModerationRejected, "moderation service unavailable", false)
			return
		}
		finish(gist, db.ModerationApproved, "moderation unavailable, published by policy", true)
		return
	}

	if allowed {
		finish(gist, db.ModerationApproved, reason, true)
		return
	}
	finish(gist, db.ModerationRejected, reason, false)
}

func finish(gist *db.Gist, status db.ModerationStatus, reason string, publish bool) {
	if err := gist.SetModerationResult(status, reason, publish, time.Now().Unix()); err != nil {
		log.Error().Err(err).Msgf("Moderation: cannot store result for gist %d", gist.ID)
		return
	}

	gist.ModerationStatus = status
	gist.ModerationReason = reason
	if publish {
		gist.Private = db.PublicVisibility
	}
	// The visible state changed, so refresh the search document too.
	gist.AddInIndex()

	if publish {
		log.Info().Msgf("Moderation: gist %d approved", gist.ID)
	} else {
		log.Info().Msgf("Moderation: gist %d kept private: %s", gist.ID, reason)
	}
}

func review(gist *db.Gist) (bool, string, error) {
	endpoint := settings.ModerationEndpoint()
	if endpoint == "" {
		return false, "", fmt.Errorf("moderation endpoint is not configured")
	}
	apiKey := settings.ModerationAPIKey()
	model := settings.ModerationModel()
	if model == "" {
		model = "gpt-4o-mini"
	}

	payload := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: settings.ModerationPrompt()},
			{Role: "user", Content: document(gist)},
		},
		Temperature: 0,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return false, "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), settings.ModerationTimeout())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, "", fmt.Errorf("moderation endpoint returned %s: %s", resp.Status, truncate(string(raw), 200))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return false, "", fmt.Errorf("cannot decode moderation response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return false, "", fmt.Errorf("moderation response has no choices")
	}

	return parseVerdict(parsed.Choices[0].Message.Content)
}

// parseVerdict tolerates models that wrap the JSON in prose or a Markdown code
// fence, which is common in practice even when the prompt asks for JSON only.
func parseVerdict(content string) (bool, string, error) {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return false, "", fmt.Errorf("moderation response contains no JSON object")
	}

	var v verdict
	if err := json.Unmarshal([]byte(content[start:end+1]), &v); err != nil {
		return false, "", fmt.Errorf("cannot decode moderation verdict: %w", err)
	}
	return v.Allowed, strings.TrimSpace(v.Reason), nil
}

func document(gist *db.Gist) string {
	var b strings.Builder
	b.WriteString("Title: ")
	b.WriteString(gist.Title)
	b.WriteString("\nDescription: ")
	b.WriteString(gist.Description)
	b.WriteString("\n\n")

	total := 0
	if files, _, err := gist.Files("HEAD", true); err == nil {
		for _, file := range files {
			header := "\n----- " + file.Filename + " -----\n"
			if total+len(header) >= maxDocumentBytes {
				break
			}
			b.WriteString(header)
			total += len(header)

			content := file.Content
			remaining := maxDocumentBytes - total
			if remaining <= 0 {
				break
			}
			if len(content) > remaining {
				content = truncate(content, remaining)
			}
			b.WriteString(content)
			b.WriteString("\n")
			total += len(content)
		}
	}

	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	// Do not cut a multi-byte character (Chinese, emoji) in half.
	for max > 0 && !utf8.RuneStart(s[max]) {
		max--
	}
	return s[:max] + "…"
}
