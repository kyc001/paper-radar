package scoring

import (
	"testing"

	"github.com/kyc001/paper-radar/internal/model"
)

func TestScoreTextCountsKeywordFrequencyCaseInsensitive(t *testing.T) {
	text := "LLM agents improve retrieval. llm systems can summarize. Agent design matters."
	keywords := []string{"llm", "agent"}

	score := ScoreText(text, keywords)
	if score != 4 {
		t.Fatalf("expected score 4, got %d", score)
	}
}

func TestFilterMinScore(t *testing.T) {
	papers := []model.ScoredPaper{
		{Score: 0},
		{Score: 2},
		{Score: 5},
	}

	filtered := FilterMinScore(papers, 2)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 papers, got %d", len(filtered))
	}
}
