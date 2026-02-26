package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kyc001/paper-radar/internal/model"
	"github.com/kyc001/paper-radar/internal/state"
)

func TestRunDigestTopNLeavesRemainingPending(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")
	outDir := filepath.Join(dir, "out")

	store := state.New(statePath)
	seed := state.FileState{
		SeenIDs: map[string]bool{"a": true, "b": true, "c": true},
		Pending: []model.ScoredPaper{
			{Paper: model.Paper{ID: "a", Title: "A", Summary: "s"}, Score: 10, Topics: []string{"t1"}},
			{Paper: model.Paper{ID: "b", Title: "B", Summary: "s"}, Score: 8, Topics: []string{"t1"}},
			{Paper: model.Paper{ID: "c", Title: "C", Summary: "s"}, Score: 6, Topics: []string{"t2"}},
		},
	}
	if err := store.Save(seed); err != nil {
		t.Fatalf("save seed: %v", err)
	}

	path, count, err := RunDigest(DigestOptions{
		StatePath: statePath,
		OutputDir: outDir,
		Date:      time.Date(2026, 2, 26, 0, 0, 0, 0, time.UTC),
		TopN:      2,
	})
	if err != nil {
		t.Fatalf("RunDigest failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count=2, got %d", count)
	}

	content := mustReadFile(t, path)
	if !strings.Contains(content, "## 1. A") || !strings.Contains(content, "## 2. B") {
		t.Fatalf("digest content missing expected papers: %s", content)
	}
	if strings.Contains(content, "## 3. C") || strings.Contains(content, "## 1. C") {
		t.Fatalf("digest should not include C when top=2")
	}

	after, err := store.Load()
	if err != nil {
		t.Fatalf("load state after digest: %v", err)
	}
	if len(after.Pending) != 1 {
		t.Fatalf("expected 1 pending after digest, got %d", len(after.Pending))
	}
	if after.Pending[0].Paper.ID != "c" {
		t.Fatalf("expected remaining pending paper c, got %s", after.Pending[0].Paper.ID)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return string(b)
}
