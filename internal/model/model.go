package model

import "time"

type Paper struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ScoredPaper struct {
	Paper  Paper    `json:"paper"`
	Score  int      `json:"score"`
	Topics []string `json:"topics"`
}
