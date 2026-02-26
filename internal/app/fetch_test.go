package app

import (
	"testing"
	"time"

	"github.com/kyc001/paper-radar/internal/config"
	"github.com/kyc001/paper-radar/internal/model"
)

func TestProcessPaperAggregatesAcrossTopics(t *testing.T) {
	t.Parallel()

	originalSeen := map[string]bool{}
	seenIDs := map[string]bool{}
	byID := map[string]model.ScoredPaper{}

	paper := model.Paper{
		ID:          "paper-1",
		Title:       "agent memory",
		Summary:     "agent memory planning",
		PublishedAt: time.Now(),
	}

	processPaper(originalSeen, seenIDs, byID, config.Topic{Name: "A", Keywords: []string{"agent"}}, paper, 1)
	processPaper(originalSeen, seenIDs, byID, config.Topic{Name: "B", Keywords: []string{"memory"}}, paper, 1)

	got, ok := byID[paper.ID]
	if !ok {
		t.Fatalf("paper should be queued")
	}
	if len(got.Topics) != 2 {
		t.Fatalf("expected 2 topics, got %v", got.Topics)
	}
	if got.Score != 4 {
		t.Fatalf("expected aggregated score 4, got %d", got.Score)
	}
}

func TestProcessPaperRespectsMinScoreButMarksSeen(t *testing.T) {
	t.Parallel()

	originalSeen := map[string]bool{}
	seenIDs := map[string]bool{}
	byID := map[string]model.ScoredPaper{}

	paper := model.Paper{ID: "paper-2", Title: "weak match", Summary: "just one keyword mention"}
	processPaper(originalSeen, seenIDs, byID, config.Topic{Name: "A", Keywords: []string{"keyword"}}, paper, 2)

	if _, ok := byID[paper.ID]; ok {
		t.Fatalf("paper should not pass min-score threshold")
	}
	if !seenIDs[paper.ID] {
		t.Fatalf("paper should still be marked as seen")
	}
}

func TestProcessPaperSkipsOriginalSeen(t *testing.T) {
	t.Parallel()

	originalSeen := map[string]bool{"paper-3": true}
	seenIDs := map[string]bool{"paper-3": true}
	byID := map[string]model.ScoredPaper{}

	paper := model.Paper{ID: "paper-3", Title: "agent", Summary: "agent"}
	processPaper(originalSeen, seenIDs, byID, config.Topic{Name: "A", Keywords: []string{"agent"}}, paper, 1)

	if len(byID) != 0 {
		t.Fatalf("already-seen paper should be skipped")
	}
}
