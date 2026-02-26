package scoring

import (
	"sort"
	"strings"

	"github.com/kyc001/paper-radar/internal/model"
)

func ScorePaper(paper model.Paper, keywords []string) int {
	content := paper.Title + " " + paper.Summary
	return ScoreText(content, keywords)
}

func ScoreText(text string, keywords []string) int {
	if len(keywords) == 0 {
		return 0
	}

	content := strings.ToLower(text)
	total := 0
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(strings.ToLower(keyword))
		if keyword == "" {
			continue
		}
		total += strings.Count(content, keyword)
	}

	return total
}

func FilterMinScore(papers []model.ScoredPaper, minScore int) []model.ScoredPaper {
	filtered := make([]model.ScoredPaper, 0, len(papers))
	for _, paper := range papers {
		if paper.Score >= minScore {
			filtered = append(filtered, paper)
		}
	}
	return filtered
}

func SortByScore(papers []model.ScoredPaper) {
	sort.Slice(papers, func(i, j int) bool {
		if papers[i].Score == papers[j].Score {
			return papers[i].Paper.PublishedAt.After(papers[j].Paper.PublishedAt)
		}
		return papers[i].Score > papers[j].Score
	})
}
