package main

import "testing"

func TestResolveWebhookPrecedence(t *testing.T) {
	t.Setenv("PAPER_RADAR_FEISHU_WEBHOOK", "env")

	if got := resolveWebhook("cli", "cfg"); got != "cli" {
		t.Fatalf("expected cli to win, got %q", got)
	}
	if got := resolveWebhook("", "cfg"); got != "cfg" {
		t.Fatalf("expected cfg to win, got %q", got)
	}
	if got := resolveWebhook("", ""); got != "env" {
		t.Fatalf("expected env fallback, got %q", got)
	}
}
