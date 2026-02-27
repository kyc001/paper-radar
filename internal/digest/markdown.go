package digest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kyc001/paper-radar/internal/model"
)

var (
	qSectionRe = regexp.MustCompile(`Q(\d+)\s*[:：]\s*`)

	// Patterns for reformatting flat (single-line) content
	headingRe      = regexp.MustCompile(`\s+(#{2,4}\s+)`)
	boldLabelRe    = regexp.MustCompile(`([^.\n])\s+(\*\*[^*]{2,60}\*\*\s*[:：])`) // not after "N. "
	boldStandRe    = regexp.MustCompile(`\s+(\*\*[^*]{2,40}\*\*)\s+(-)`)           // standalone bold before list
	listItemRe     = regexp.MustCompile(`([^\n])\s+(- \*\*)`)
	plainListRe    = regexp.MustCompile(`([^\n-])\s+(- [^*\n])`)
	numListRe      = regexp.MustCompile(`([^\n#])\s+(\d+\.\s+)`) // don't break ### 1. headings
	mathBlockRe    = regexp.MustCompile(`([^\$\n])\s*(\$\$)`)
	tableRowRe     = regexp.MustCompile(`([^\n|])\s*(\|[^|\n]+\|[^|\n]+\|)`)
	excessNLRe     = regexp.MustCompile(`\n{3,}`)
	brokenNumBold  = regexp.MustCompile(`(\d+\.)\s*\n+(\*\*)`)  // fix "1.\n\n**text**"
	brokenDashBold = regexp.MustCompile(`(- )\s*\n+(\*\*)`)     // fix "- \n\n**text**"
)

// Q section default titles (Chinese)
var qTitles = map[string]string{
	"Q1": "这篇论文试图解决什么问题？",
	"Q2": "有哪些相关研究？",
	"Q3": "论文如何解决这个问题？",
	"Q4": "论文做了哪些实验？",
	"Q5": "有什么可以进一步探索的点？",
	"Q6": "总结一下论文的主要内容",
}

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

	fmt.Fprintf(&builder, "> %d papers | Auto-formatted\n\n", len(papers))

	for i, paper := range papers {
		writePaperMarkdown(&builder, i+1, paper)
		if i < len(papers)-1 {
			builder.WriteString("---\n\n")
		}
	}

	return builder.String()
}

func writePaperMarkdown(builder *strings.Builder, num int, paper model.ScoredPaper) {
	// Title
	fmt.Fprintf(builder, "## %d. %s\n\n", num, paper.Paper.Title)

	// Metadata table
	builder.WriteString("| Field | Value |\n")
	builder.WriteString("|-------|-------|\n")
	fmt.Fprintf(builder, "| Score | %d |\n", paper.Score)
	if len(paper.Topics) > 0 {
		fmt.Fprintf(builder, "| Topics | %s |\n", strings.Join(paper.Topics, ", "))
	}
	if paper.Paper.URL != "" {
		arxivID := paper.Paper.URL
		if idx := strings.LastIndex(arxivID, "/"); idx >= 0 {
			arxivID = arxivID[idx+1:]
		}
		fmt.Fprintf(builder, "| URL | [%s](%s) |\n", arxivID, paper.Paper.URL)
	}
	if !paper.Paper.PublishedAt.IsZero() {
		fmt.Fprintf(builder, "| Published | %s |\n", paper.Paper.PublishedAt.Format("2006-01-02"))
	}
	builder.WriteString("\n")

	// Format summary with Q sections
	summary := paper.Paper.Summary
	if summary == "" {
		builder.WriteString("*(No summary available)*\n\n")
		return
	}

	sections := splitQSections(summary)
	if len(sections) == 0 {
		// No Q sections found — output as-is
		builder.WriteString(summary)
		builder.WriteString("\n\n")
		return
	}

	for _, qKey := range []string{"Q1", "Q2", "Q3", "Q4", "Q5", "Q6"} {
		content, ok := sections[qKey]
		if !ok {
			continue
		}

		title := qTitles[qKey]
		// Strip the default title from content if it starts with it
		if title != "" && strings.HasPrefix(content, title) {
			content = strings.TrimSpace(content[len(title):])
		}

		fmt.Fprintf(builder, "### %s: %s\n\n", qKey, title)

		// If content is flat (no newlines), reformat it
		if strings.Count(content, "\n") < 3 && len(content) > 200 {
			content = reformatFlatContent(content)
		}

		builder.WriteString(content)
		builder.WriteString("\n\n")
	}
	// Skip Q7 (Kimi promo)
}

// reformatFlatContent inserts newlines before markdown structural elements
// in single-line content (from old-format state data where stripHTML collapsed newlines).
func reformatFlatContent(text string) string {
	// Headings: ## N. or ### N.
	text = headingRe.ReplaceAllString(text, "\n\n$1")

	// Bold sub-headers: **label**：
	text = boldLabelRe.ReplaceAllString(text, "\n\n$1")

	// Bold standalone labels before list items: **label** - item
	text = boldStandRe.ReplaceAllString(text, "\n\n$1\n$2")

	// List items: - **bold**
	text = listItemRe.ReplaceAllString(text, "$1\n$2")

	// Plain list items: - text
	text = plainListRe.ReplaceAllString(text, "$1\n$2")

	// Numbered list items: 1. text (but not after #)
	text = numListRe.ReplaceAllString(text, "$1\n\n$2")

	// Math blocks: $$...$$ on own lines
	text = mathBlockRe.ReplaceAllString(text, "$1\n\n$$")
	// Ensure closing $$ gets a newline after it
	text = regexp.MustCompile(`(\$\$)([^\n$])`).ReplaceAllString(text, "$$\n\n$2")

	// Table rows
	text = tableRowRe.ReplaceAllString(text, "$1\n$2")

	// Post-fix: rejoin broken patterns where number/dash got separated from bold
	text = brokenNumBold.ReplaceAllString(text, "$1 $2")
	text = brokenDashBold.ReplaceAllString(text, "$1$2")

	// Clean up excessive newlines
	text = excessNLRe.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// splitQSections splits a summary into Q1-Q7 sections.
func splitQSections(summary string) map[string]string {
	indices := qSectionRe.FindAllStringSubmatchIndex(summary, -1)
	nums := qSectionRe.FindAllStringSubmatch(summary, -1)

	if len(nums) == 0 {
		return nil
	}

	sections := make(map[string]string, len(nums))
	for i, match := range nums {
		qNum := match[1]
		// Content starts after this match, ends at next match or end
		startIdx := indices[i][1] // end of the match
		var endIdx int
		if i+1 < len(indices) {
			endIdx = indices[i+1][0] // start of next match
		} else {
			endIdx = len(summary)
		}
		content := strings.TrimSpace(summary[startIdx:endIdx])
		sections["Q"+qNum] = content
	}

	return sections
}
