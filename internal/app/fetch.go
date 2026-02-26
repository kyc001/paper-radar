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
	fetchedCount := 0

	for _, topic := range cfg.Topics {
		maxResults := cfg.EffectiveMaxResults(topic, opts.MaxResults)
		query := cfg.TopicQuery(topic)

		papers, err := client.Fetch(ctx, query, maxResults)
		if err != nil {
			return FetchResult{}, fmt.Errorf("fetch topic %q: %w", topic.Name, err)
		}

		fetchedCount += len(papers)
		for _, paper := range papers {
			if st.SeenIDs[paper.ID] {
				continue
			}

			// Mark as seen immediately to avoid repeated processing in later runs.
			st.SeenIDs[paper.ID] = true

			score := scoring.ScorePaper(paper, topic.Keywords)
			if score <= 0 {
				continue
			}

			existing, ok := newByID[paper.ID]
			if !ok {
				newByID[paper.ID] = model.ScoredPaper{
					Paper:  paper,
					Score:  score,
					Topics: []string{topic.Name},
				}
				continue
			}

			existing.Score += score
			existing.Topics = appendIfMissing(existing.Topics, topic.Name)
			newByID[paper.ID] = existing
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
