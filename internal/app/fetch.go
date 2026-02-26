package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/kyc001/paper-radar/internal/arxiv"
	"github.com/kyc001/paper-radar/internal/config"
	"github.com/kyc001/paper-radar/internal/model"
	"github.com/kyc001/paper-radar/internal/scoring"
	"github.com/kyc001/paper-radar/internal/state"
)

const DefaultStatePath = ".paper-radar/state.json"

type FetchOptions struct {
	ConfigPath string
	StatePath  string
	MaxResults int
	MinScore   int
}

type FetchResult struct {
	Fetched int
	Queued  int
	Topics  int
}

func RunFetch(ctx context.Context, opts FetchOptions) (FetchResult, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return FetchResult{}, fmt.Errorf("load config: %w", err)
	}

	store := state.New(defaultStatePath(opts.StatePath))
	st, err := store.Load()
	if err != nil {
		return FetchResult{}, fmt.Errorf("load state: %w", err)
	}

	client := arxiv.NewClient()
	newByID := make(map[string]model.ScoredPaper)
	originalSeen := cloneSeen(st.SeenIDs)
	fetchedCount := 0

	for _, topic := range cfg.Topics {
		maxResults := cfg.EffectiveMaxResults(topic, opts.MaxResults)
		minScore := cfg.EffectiveMinScore(topic, opts.MinScore)
		query := cfg.TopicQuery(topic)

		papers, err := client.Fetch(ctx, query, maxResults)
		if err != nil {
			return FetchResult{}, fmt.Errorf("fetch topic %q: %w", topic.Name, err)
		}

		fetchedCount += len(papers)
		for _, paper := range papers {
			processPaper(originalSeen, st.SeenIDs, newByID, topic, paper, minScore)
		}
	}

	newPapers := mapToSortedSlice(newByID)
	st.Pending = append(st.Pending, newPapers...)

	if err := store.Save(st); err != nil {
		return FetchResult{}, fmt.Errorf("save state: %w", err)
	}

	return FetchResult{
		Fetched: fetchedCount,
		Queued:  len(newPapers),
		Topics:  len(cfg.Topics),
	}, nil
}

func processPaper(originalSeen map[string]bool, seenIDs map[string]bool, byID map[string]model.ScoredPaper, topic config.Topic, paper model.Paper, minScore int) {
	if originalSeen[paper.ID] {
		return
	}

	score := scoring.ScorePaper(paper, topic.Keywords)
	if score >= minScore {
		existing, ok := byID[paper.ID]
		if !ok {
			byID[paper.ID] = model.ScoredPaper{
				Paper:  paper,
				Score:  score,
				Topics: []string{topic.Name},
			}
		} else {
			existing.Score += score
			existing.Topics = appendIfMissing(existing.Topics, topic.Name)
			byID[paper.ID] = existing
		}
	}

	// Mark as seen even if it didn't pass threshold, so the next run won't reprocess it.
	seenIDs[paper.ID] = true
}

func cloneSeen(seen map[string]bool) map[string]bool {
	cloned := make(map[string]bool, len(seen))
	for id, ok := range seen {
		cloned[id] = ok
	}
	return cloned
}

func mapToSortedSlice(byID map[string]model.ScoredPaper) []model.ScoredPaper {
	papers := make([]model.ScoredPaper, 0, len(byID))
	for _, paper := range byID {
		papers = append(papers, paper)
	}

	sort.Slice(papers, func(i, j int) bool {
		if papers[i].Score == papers[j].Score {
			return papers[i].Paper.PublishedAt.After(papers[j].Paper.PublishedAt)
		}
		return papers[i].Score > papers[j].Score
	})

	return papers
}

func appendIfMissing(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func defaultStatePath(statePath string) string {
	if statePath != "" {
		return statePath
	}
	return DefaultStatePath
}
