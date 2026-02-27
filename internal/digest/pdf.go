package digest

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/kyc001/paper-radar/internal/model"
)

func WritePDF(outputDir string, date time.Time, papers []model.ScoredPaper) (string, error) {
	filename := date.Format("2006-01-02") + ".pdf"
	path := filepath.Join(outputDir, filename)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Title
	title := fmt.Sprintf("Paper Radar Digest - %s", date.Format("2006-01-02"))
	pdf.Cell(0, 10, title)
	pdf.Ln(12)

	if len(papers) == 0 {
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(0, 10, "No new papers matched the configured keywords.")
	} else {
		pdf.SetFont("Arial", "", 11)
		lineHeight := 6.0
		maxWidth := 190.0

		for i, paper := range papers {
			// Paper number and title
			pdf.SetFont("Arial", "B", 12)
			header := fmt.Sprintf("%d. %s", i+1, paper.Paper.Title)
			pdf.MultiCell(0, lineHeight, "", "", truncateString(header, 180), false)
			pdf.Ln(2)

			// Metadata
			pdf.SetFont("Arial", "", 10)
			pdf.Cell(0, lineHeight, fmt.Sprintf("Score: %d | Topics: %s", paper.Score, joinStrings(paper.Topics, ", ")))
			pdf.Ln(lineHeight)
			pdf.Cell(0, lineHeight, fmt.Sprintf("ID: %s", paper.Paper.ID))
			pdf.Ln(lineHeight)

			// URL - use monospace for URL
			pdf.SetFont("Courier", "", 9)
			pdf.Cell(0, lineHeight, fmt.Sprintf("URL: %s", paper.Paper.URL))
			pdf.Ln(lineHeight)

			if !paper.Paper.PublishedAt.IsZero() {
				pdf.SetFont("Arial", "", 10)
				pdf.Cell(0, lineHeight, fmt.Sprintf("Published: %s", paper.Paper.PublishedAt.Format("2006-01-02")))
				pdf.Ln(lineHeight)
			}
			pdf.Ln(2)

			// Summary with word wrapping
			pdf.SetFont("Arial", "", 10)
			summary := paper.Paper.Summary
			for _, line := range splitText(summary, int(maxWidth)) {
				pdf.MultiCell(0, lineHeight, "", "", line, false)
			}

			// Add space between papers
			pdf.Ln(5)

			// Add page break if needed
			if pdf.GetY() > 250 {
				pdf.AddPage()
			}
		}
	}

	if err := pdf.OutputFileAndClose(path); err != nil {
		return "", err
	}

	return path, nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

// splitText splits text into lines that fit within maxWidth characters
func splitText(text string, maxWidth int) []string {
	var lines []string
	words := splitWords(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= maxWidth {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// splitWords splits text into words (handles newlines)
func splitWords(text string) []string {
	var words []string
	for _, line := range splitByNewline(text) {
		for _, word := range splitBySpace(line) {
			if word != "" {
				words = append(words, word)
			}
		}
	}
	return words
}

func splitByNewline(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' || r == '\r' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitBySpace(s string) []string {
	var words []string
	current := ""
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
