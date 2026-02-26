package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultMaxChunkChars = 2800

type FeishuWebhook struct {
	httpClient *http.Client
}

func NewFeishuWebhook() *FeishuWebhook {
	return &FeishuWebhook{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (f *FeishuWebhook) SendText(ctx context.Context, webhookURL, text string) error {
	if strings.TrimSpace(webhookURL) == "" {
		return fmt.Errorf("webhook url is empty")
	}

	payload := map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": text,
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("feishu webhook status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}

func (f *FeishuWebhook) SendLongText(ctx context.Context, webhookURL, text string, maxChunkChars int) error {
	if maxChunkChars <= 0 {
		maxChunkChars = defaultMaxChunkChars
	}
	parts := splitText(text, maxChunkChars)
	for i, part := range parts {
		msg := part
		if len(parts) > 1 {
			msg = fmt.Sprintf("[%d/%d]\n%s", i+1, len(parts), part)
		}
		if err := f.SendText(ctx, webhookURL, msg); err != nil {
			return err
		}
	}
	return nil
}

func splitText(text string, maxChunkChars int) []string {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return []string{clean}
	}

	runes := []rune(clean)
	if len(runes) <= maxChunkChars {
		return []string{clean}
	}

	parts := make([]string, 0)
	for len(runes) > maxChunkChars {
		split := maxChunkChars
		for i := maxChunkChars; i > maxChunkChars/2; i-- {
			if runes[i-1] == '\n' {
				split = i
				break
			}
		}
		parts = append(parts, strings.TrimSpace(string(runes[:split])))
		runes = runes[split:]
	}

	tail := strings.TrimSpace(string(runes))
	if tail != "" {
		parts = append(parts, tail)
	}
	if len(parts) == 0 {
		return []string{clean}
	}
	return parts
}
