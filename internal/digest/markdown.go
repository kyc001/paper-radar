package digest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kyc001/paper-radar/internal/model"
)

func WriteDaily(outputDir string, date time.Time, papers []model.ScoredPaper) (string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}

	filename := date.Format("2006-01-02") + ".md"
	path := filepath.Join(outputDir, filename)

	content := BuildMarkdown(date, papers)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}

	return path, nil
}

func BuildMarkdown(date time.Time, papers []model.ScoredPaper) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "# Paper Radar Digest %s\n\n", date.Format("2006-01-02"))
	if len(papers) == 0 {
		builder.WriteString("No new papers matched the configured keywords.\n")
		return builder.String()
	}

	for i, paper := range papers {
		fmt.Fprintf(&builder, "## %d. %s\n\n", i+1, paper.Paper.Title)
		fmt.Fprintf(&builder, "- Score: %d\n", paper.Score)
		fmt.Fprintf(&builder, "- Topics: %s\n", strings.Join(paper.Topics, ", "))
		fmt.Fprintf(&builder, "- ID: `%s`\n", paper.Paper.ID)
		fmt.Fprintf(&builder, "- URL: %s\n", paper.Paper.URL)
		if !paper.Paper.PublishedAt.IsZero() {
			fmt.Fprintf(&builder, "- Published: %s\n", paper.Paper.PublishedAt.Format("2006-01-02"))
		}
		builder.WriteString("\n")
		builder.WriteString(paper.Paper.Summary)
		builder.WriteString("\n\n")
	}

	return builder.String()
}
