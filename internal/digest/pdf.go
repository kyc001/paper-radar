package digest

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/kyc001/paper-radar/internal/model"
)

func WritePDF(outputDir string, date time.Time, papers []model.ScoredPaper) (string, error) {
	filename := date.Format("2006-01-02") + ".pdf"
	path := filepath.Join(outputDir, filename)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false) // 禁用压缩，便于调试
	
	// 使用内置字体（仅支持 ASCII）
	pdf.AddPage()
	pdf.SetFont("Courier", "B", 12)
	
	// Title
	title := fmt.Sprintf("Paper Radar Digest - %s", date.Format("2006-01-02"))
	pdf.Cell(0, 10, title)
	pdf.Ln(12)

	if len(papers) == 0 {
		pdf.SetFont("Courier", "", 11)
		pdf.Cell(0, 10, "No new papers matched the configured keywords.")
	} else {
		pdf.SetFont("Courier", "", 9)
		lineHeight := 4.5

		for i, paper := range papers {
			// Page break if needed
			if pdf.GetY() > 270 {
				pdf.AddPage()
			}

			// Header
			pdf.SetFont("Courier", "B", 10)
			pdf.Cell(0, lineHeight, fmt.Sprintf("%d. %s", i+1, truncateString(paper.Paper.Title, 70)))
			pdf.Ln(lineHeight)
			
			// Metadata
			pdf.SetFont("Courier", "", 8)
			pdf.Cell(0, lineHeight, fmt.Sprintf("Score: %d", paper.Score))
			pdf.Ln(lineHeight)
			pdf.Cell(0, lineHeight, fmt.Sprintf("URL: %s", paper.Paper.URL))
			pdf.Ln(lineHeight * 1.5)

			// Summary (first 1500 chars)
			summary := truncateString(paper.Paper.Summary, 1500)
			lines := wrapText(summary, 90)
			for _, line := range lines {
				pdf.Cell(0, lineHeight, line)
				pdf.Ln(lineHeight)
			}

			// Separator
			pdf.Ln(3)
			pdf.Cell(0, 0, strings.Repeat("-", 180))
			pdf.Ln(5)
		}
	}

	err := pdf.OutputFileAndClose(path)
	return path, err
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func wrapText(text string, maxWidth int) []string {
	var lines []string
	words := strings.Fields(text)
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
