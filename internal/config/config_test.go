package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParsesMinScoreAndMaxResults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `max_results: 20
min_score: 2
feishu_webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/test"
topics:
  - name: "Topic A"
    source: "paperscool"
    query: "cs.CV"
    max_results: 10
    min_score: 4
    kimi_summary: true
    keywords:
      - "video"
      - "training-free"
  - name: "Topic B"
    keywords:
      - "memory"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.MaxResults != 20 {
		t.Fatalf("expected MaxResults=20, got %d", cfg.MaxResults)
	}
	if cfg.MinScore != 2 {
		t.Fatalf("expected MinScore=2, got %d", cfg.MinScore)
	}
	if cfg.FeishuWebhook == "" {
		t.Fatalf("expected FeishuWebhook parsed")
	}
	if len(cfg.Topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(cfg.Topics))
	}
	if cfg.Topics[0].MaxResults != 10 || cfg.Topics[0].MinScore != 4 {
		t.Fatalf("topic A max/min parse mismatch: max=%d min=%d", cfg.Topics[0].MaxResults, cfg.Topics[0].MinScore)
	}
	if cfg.Topics[0].Source != "paperscool" {
		t.Fatalf("expected topic source paperscool, got %q", cfg.Topics[0].Source)
	}
	if !cfg.Topics[0].KimiSummary {
		t.Fatalf("expected topic kimi_summary=true")
	}
	if cfg.Topics[1].Source != "arxiv" {
		t.Fatalf("default source should be arxiv, got %q", cfg.Topics[1].Source)
	}
}

func TestEffectiveMinScorePrecedence(t *testing.T) {
	t.Parallel()

	cfg := Config{MinScore: 2}
	topic := Topic{MinScore: 3}

	if got := cfg.EffectiveMinScore(topic, 5); got != 5 {
		t.Fatalf("cli override should win, got %d", got)
	}
	if got := cfg.EffectiveMinScore(topic, 0); got != 3 {
		t.Fatalf("topic min_score should win, got %d", got)
	}

	cfg = Config{MinScore: 2}
	topic = Topic{}
	if got := cfg.EffectiveMinScore(topic, 0); got != 2 {
		t.Fatalf("config min_score should win, got %d", got)
	}

	cfg = Config{}
	if got := cfg.EffectiveMinScore(topic, 0); got != 1 {
		t.Fatalf("default min_score should be 1, got %d", got)
	}
}
