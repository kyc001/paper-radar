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

func WriteHTML(outputDir string, date time.Time, papers []model.ScoredPaper) (string, error) {
	filename := date.Format("2006-01-02") + ".html"
	path := filepath.Join(outputDir, filename)

	var builder strings.Builder

	builder.WriteString(`<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Paper Radar Digest</title>
<style>
:root {
  --color-primary: #2563eb;
  --color-secondary: #64748b;
  --color-accent: #f59e0b;
  --color-success: #10b981;
  --color-bg: #ffffff;
  --color-bg-alt: #f8fafc;
  --color-border: #d1d5db;
  --color-text: #1f2937;
  --color-text-light: #6b7280;
}

@page {
  size: A4;
  margin: 20mm 16mm;
  @top-center { content: none; }
  @top-left { content: none; }
  @top-right { content: none; }
  @bottom-center { content: none; }
  @bottom-left { content: none; }
  @bottom-right { content: none; }
}

* { box-sizing: border-box; -webkit-print-color-adjust: exact; print-color-adjust: exact; }

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans SC", "Microsoft YaHei", sans-serif;
  font-size: 11pt;
  line-height: 1.8;
  color: var(--color-text);
  background: var(--color-bg);
  margin: 0;
  padding: 0;
}

/* ========== Document Header ========== */
.doc-header {
  text-align: center;
  padding: 48px 0;
  border-bottom: 4px solid var(--color-primary);
  margin-bottom: 48px;
  page-break-after: always;
}

.doc-header h1 {
  font-size: 32pt;
  color: var(--color-primary);
  margin: 0 0 12px 0;
  font-weight: 800;
  letter-spacing: -0.5px;
}

.doc-header .subtitle {
  font-size: 16pt;
  color: var(--color-text-light);
  margin: 0 0 16px 0;
  font-weight: 400;
}

.doc-header .meta-info {
  font-size: 11pt;
  color: var(--color-text-light);
  margin: 0;
}

/* ========== Table of Contents ========== */
.toc {
  background: var(--color-bg-alt);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  padding: 32px;
  margin-bottom: 48px;
}

.toc h2 {
  font-size: 18pt;
  color: var(--color-text);
  margin: 0 0 24px 0;
  padding: 0;
  border: none;
  background: none;
  font-weight: 700;
}

.toc ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.toc li {
  margin: 12px 0;
  padding: 8px 0;
  border-bottom: 1px dashed var(--color-border);
}

.toc li:last-child {
  border-bottom: none;
}

.toc a {
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 600;
  font-size: 12pt;
}

.toc a:hover {
  text-decoration: underline;
}

/* ========== Paper Article ========== */
.paper {
  margin-bottom: 48px;
  page-break-before: always;
}

.paper:first-child {
  page-break-before: avoid;
}

.paper__head {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  border-left: 6px solid var(--color-primary);
  border-radius: 12px;
  padding: 32px;
  margin-bottom: 32px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.paper__title {
  font-size: 20pt;
  font-weight: 800;
  color: var(--color-text);
  margin: 0 0 16px 0;
  line-height: 1.3;
  letter-spacing: -0.3px;
}

.paper__meta {
  font-size: 10.5pt;
  color: var(--color-text-light);
  margin: 12px 0;
  word-break: break-all;
}

.paper__meta a {
  color: var(--color-primary);
  text-decoration: none;
  border-bottom: 1px dotted var(--color-primary);
}

.paper__meta a:hover {
  border-bottom-style: solid;
}

.paper__meta a::after {
  content: " (" attr(href) ")";
  font-size: 9pt;
  color: var(--color-text-light);
}

.paper__score {
  display: inline-block;
  background: var(--color-success);
  color: white;
  padding: 4px 14px;
  border-radius: 16px;
  font-weight: 700;
  font-size: 10pt;
  margin-left: 12px;
  box-shadow: 0 2px 4px rgba(16,185,129,0.3);
}

.paper__takeaway {
  background: white;
  border: 2px solid var(--color-border);
  border-radius: 8px;
  padding: 20px;
  margin-top: 24px;
}

.paper__takeaway-title {
  font-size: 10pt;
  font-weight: 700;
  color: var(--color-text-light);
  margin: 0 0 8px 0;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.paper__takeaway-content {
  font-size: 11pt;
  color: var(--color-text);
  margin: 0;
  line-height: 1.7;
  font-style: italic;
}

/* ========== Q&A Section ========== */
.paper__body {
  padding: 0;
}

.qa {
  margin-bottom: 32px;
  page-break-inside: avoid;
}

.qa__q {
  font-size: 14pt;
  font-weight: 700;
  color: var(--color-primary);
  background: linear-gradient(90deg, #eff6ff 0%, #f8fafc 100%);
  border-left: 5px solid var(--color-primary);
  padding: 16px 20px;
  margin: 0 0 20px 0;
  border-radius: 0 8px 8px 0;
  letter-spacing: -0.2px;
}

.qa__a {
  padding: 0 8px;
}

.qa__a p {
  margin: 14px 0;
  text-align: justify;
  line-height: 1.9;
}

.qa__a ul {
  margin: 14px 0;
  padding-left: 32px;
}

.qa__a li {
  margin: 10px 0;
  line-height: 1.8;
}

/* ========== Tables ========== */
.table-wrapper {
  margin: 24px 0;
  overflow-x: auto;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 10pt;
  page-break-inside: auto;
}

tr {
  page-break-inside: avoid;
  page-break-after: auto;
}

th, td {
  border: 1px solid var(--color-border);
  padding: 14px 16px;
  text-align: left;
  vertical-align: top;
  line-height: 1.6;
}

th {
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
  font-weight: 700;
  color: var(--color-text);
  font-size: 10.5pt;
}

tr:nth-child(even) td {
  background: #f8fafc;
}

tr:hover td {
  background: #f1f5f9;
}

/* ========== Typography ========== */
p {
  margin: 14px 0;
}

strong {
  font-weight: 700;
  color: #111827;
}

code, .formula {
  background: #fef3c7;
  padding: 3px 8px;
  border-radius: 4px;
  font-family: "Courier New", "Consolas", "Monaco", monospace;
  font-size: 10pt;
  color: #92400e;
  border: 1px solid #fcd34d;
}

.formula {
  display: inline-block;
  background: #f3f4f6;
  padding: 4px 10px;
  border-radius: 6px;
  font-family: "Courier New", "Consolas", monospace;
  font-size: 10pt;
  color: #1f2937;
  border: 1px solid #d1d5db;
}

/* ========== Page Breaks ========== */
.page-break {
  page-break-before: always;
  margin-top: 48px;
  border-top: 3px dashed var(--color-border);
}

.no-break {
  page-break-inside: avoid;
}

/* ========== Print Optimizations ========== */
@media print {
  body {
    font-size: 10.5pt;
  }
  
  .paper {
    page-break-inside: avoid;
  }
  
  .qa {
    page-break-inside: avoid;
  }
  
  .table-wrapper {
    page-break-inside: avoid;
  }
  
  table {
    font-size: 9.5pt;
  }
  
  th, td {
    padding: 10px 12px;
  }
  
  a {
    text-decoration: none;
    color: #1f2937;
  }
  
  a::after {
    content: " (" attr(href) ")";
    font-size: 8.5pt;
    color: #6b7280;
    word-break: break-all;
  }
}
</style>
</head><body>
`)

	// Document Header
	builder.WriteString("<header class='doc-header'>\n")
	builder.WriteString(fmt.Sprintf("<h1>📊 Paper Radar Digest</h1>\n"))
	builder.WriteString(fmt.Sprintf("<p class='subtitle'>Weekly Research Summary</p>\n"))
	builder.WriteString(fmt.Sprintf("<p class='meta-info'>Generated: %s &nbsp;|&nbsp; Total Papers: <strong>%d</strong></p>\n", date.Format("2006-01-02"), len(papers)))
	builder.WriteString("</header>\n\n")

	// Table of Contents
	builder.WriteString("<nav class='toc'>\n")
	builder.WriteString("<h2>📑 目录 / Contents</h2>\n<ul>\n")
	for i, paper := range papers {
		title := escapeHTML(paper.Paper.Title)
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		builder.WriteString(fmt.Sprintf("<li><a href='#paper%d'>%d. %s</a></li>\n", i+1, i+1, title))
	}
	builder.WriteString("</ul>\n</nav>\n\n")

	// Papers
	for i, paper := range papers {
		if i > 0 {
			builder.WriteString("<div class='page-break'></div>\n")
		}
		
		builder.WriteString(fmt.Sprintf("<article class='paper' id='paper%d'>\n", i+1))
		
		// Paper Header
		builder.WriteString("<header class='paper__head'>\n")
		builder.WriteString(fmt.Sprintf("<h2 class='paper__title'>%d. %s</h2>\n", i+1, escapeHTML(paper.Paper.Title)))
		builder.WriteString(fmt.Sprintf("<p class='paper__meta'>🔗 <a href='%s'>%s</a> <span class='paper__score'>Score: %d</span></p>\n", paper.Paper.URL, paper.Paper.URL, paper.Score))
		
		// One-line takeaway
		takeaway := extractTakeaway(paper.Paper.Summary)
		if takeaway != "" {
			builder.WriteString("<div class='paper__takeaway'>\n")
			builder.WriteString("<p class='paper__takeaway-title'>💡 One-line Takeaway</p>\n")
			builder.WriteString(fmt.Sprintf("<p class='paper__takeaway-content'>%s</p>\n", escapeHTML(takeaway)))
			builder.WriteString("</div>\n")
		}
		builder.WriteString("</header>\n\n")
		
		// Paper Body (Q&A)
		builder.WriteString("<div class='paper__body'>\n")
		builder.WriteString(markdownToQA(paper.Paper.Summary))
		builder.WriteString("</div>\n")
		
		builder.WriteString("</article>\n\n")
	}

	builder.WriteString("</body></html>")

	content := builder.String()
	err := os.WriteFile(path, []byte(content), 0644)
	return path, err
}

// extractTakeaway extracts the first sentence as a one-line takeaway
func extractTakeaway(summary string) string {
	for i, char := range summary {
		if char == '.' || char == '!' || char == '。' {
			if i > 20 && i < 200 {
				return summary[:i+1]
			}
		}
	}
	if len(summary) > 150 {
		return summary[:150] + "..."
	}
	return summary
}

// markdownToQA converts Markdown summary to Q&A structure
func markdownToQA(text string) string {
	// Step 1: Normalize Unicode
	text = normalizeUnicode(text)
	
	// Step 2: Pre-process: add newlines before Markdown elements
	text = strings.ReplaceAll(text, "###", "\n###")
	text = regexp.MustCompile(`([^#\n])(##)`).ReplaceAllString(text, "$1\n$2")
	text = regexp.MustCompile(`([^ \n])([-*] )`).ReplaceAllString(text, "$1\n$2")
	text = regexp.MustCompile(`(Q\d+)\s*:`).ReplaceAllString(text, "\n$1:")
	
	var result strings.Builder
	
	// Step 3: Split by ### to get major sections
	sections := strings.Split(text, "###")
	
	for i, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		
		if i > 0 {
			// This is a ### section - first line is header
			lines := strings.SplitN(section, "\n", 2)
			header := strings.TrimSpace(lines[0])
			header = strings.TrimPrefix(header, "##")
			header = strings.TrimSpace(header)
			
			// Check if it's a Q# question
			if regexp.MustCompile(`^Q\d+$`).MatchString(header) {
				result.WriteString(fmt.Sprintf("<div class='qa'>\n"))
				result.WriteString(fmt.Sprintf("<h3 class='qa__q'>%s</h3>\n", processInline(header)))
				if len(lines) > 1 {
					content := strings.TrimSpace(lines[1])
					if content != "" {
						result.WriteString(fmt.Sprintf("<div class='qa__a'>%s</div>\n", processBody(content)))
					}
				}
				result.WriteString("</div>\n")
			} else {
				result.WriteString(fmt.Sprintf("<h3>%s</h3>\n", processInline(header)))
				if len(lines) > 1 {
					content := strings.TrimSpace(lines[1])
					if content != "" {
						result.WriteString(processBody(content))
					}
				}
			}
		} else {
			// First section - look for Q# questions
			result.WriteString(processBody(section))
		}
	}
	
	return result.String()
}

// normalizeUnicode normalizes Unicode characters
func normalizeUnicode(text string) string {
	var result strings.Builder
	for _, r := range text {
		// Remove soft hyphens, zero-width spaces, and other invisible characters
		if r == '\u00ad' || r == '\u200b' || r == '\u200c' || r == '\u200d' || r == '\ufeff' {
			continue
		}
		// Replace replacement character with space
		if r == '\ufffd' {
			result.WriteRune(' ')
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// processBody handles lists, tables, and paragraphs
func processBody(text string) string {
	var result strings.Builder
	lines := strings.Split(text, "\n")
	
	inList := false
	inTable := false
	var tableRows []string
	var paragraphLines []string
	
	flushParagraph := func() {
		if len(paragraphLines) > 0 {
			result.WriteString(fmt.Sprintf("<p>%s</p>\n", strings.Join(paragraphLines, " ")))
			paragraphLines = nil
		}
		if inList {
			result.WriteString("</ul>\n")
			inList = false
		}
		if inTable {
			result.WriteString(processTable(tableRows))
			tableRows = nil
			inTable = false
		}
	}
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "" {
			flushParagraph()
			continue
		}
		
		// Table row (must have at least 2 | characters)
		if strings.Count(trimmed, "|") >= 2 && !regexp.MustCompile(`^\|?[\s\-:|]+\|?$`).MatchString(trimmed) {
			if len(paragraphLines) > 0 {
				result.WriteString(fmt.Sprintf("<p>%s</p>\n", strings.Join(paragraphLines, " ")))
				paragraphLines = nil
			}
			if inList {
				result.WriteString("</ul>\n")
				inList = false
			}
			if !inTable {
				inTable = true
			}
			tableRows = append(tableRows, trimmed)
			continue
		}
		
		// List items
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if len(paragraphLines) > 0 {
				result.WriteString(fmt.Sprintf("<p>%s</p>\n", strings.Join(paragraphLines, " ")))
				paragraphLines = nil
			}
			if inTable {
				result.WriteString(processTable(tableRows))
				tableRows = nil
				inTable = false
			}
			if !inList {
				result.WriteString("<ul>\n")
				inList = true
			}
			content := strings.TrimPrefix(trimmed, "- ")
			content = strings.TrimPrefix(content, "* ")
			result.WriteString(fmt.Sprintf("<li>%s</li>\n", processInline(content)))
			continue
		}
		
		// Regular text
		if inList {
			result.WriteString("</ul>\n")
			inList = false
		}
		if inTable {
			result.WriteString(processTable(tableRows))
			tableRows = nil
			inTable = false
		}
		paragraphLines = append(paragraphLines, processInline(trimmed))
	}
	
	flushParagraph()
	return result.String()
}

// processTable converts table rows to HTML table
func processTable(rows []string) string {
	if len(rows) == 0 {
		return ""
	}
	
	var result strings.Builder
	result.WriteString("<div class='table-wrapper'>\n<table>\n")
	
	for i, row := range rows {
		row = strings.TrimSpace(row)
		row = strings.TrimPrefix(row, "|")
		row = strings.TrimSuffix(row, "|")
		cells := strings.Split(row, "|")
		
		tag := "td"
		if i == 0 {
			tag = "th"
		}
		
		result.WriteString("<tr>\n")
		for _, cell := range cells {
			cell = strings.TrimSpace(cell)
			if cell != "" {
				result.WriteString(fmt.Sprintf("<%s>%s</%s>\n", tag, processInline(cell), tag))
			}
		}
		result.WriteString("</tr>\n")
	}
	
	result.WriteString("</table>\n</div>\n")
	return result.String()
}

// processInline handles **bold**, `code`, formulas, and fixes HTML entities
func processInline(text string) string {
	// Step 1: Decode HTML entities
	text = strings.ReplaceAll(text, "&amp;#x27;", "'")
	text = strings.ReplaceAll(text, "&amp;quot;", "\"")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	
	// Step 2: Normalize Unicode (remove invisible chars)
	text = normalizeUnicode(text)
	
	// Step 3: Fix CJK spacing (add space between CJK and Latin)
	text = regexp.MustCompile(`([\p{Han}\p{Katakana}\p{Hiragana}])([a-zA-Z0-9])`).ReplaceAllString(text, "$1 $2")
	text = regexp.MustCompile(`([a-zA-Z0-9])([\p{Han}\p{Katakana}\p{Hiragana}])`).ReplaceAllString(text, "$1 $2")
	
	// Step 4: Escape HTML
	text = escapeHTML(text)
	
	// Step 5: Process formulas ($$ ... $$ or $ ... $)
	formulaRe := regexp.MustCompile(`\$\$([^$]+)\$\$`)
	text = formulaRe.ReplaceAllString(text, "<span class='formula'>$$$1$$</span>")
	
	// Step 6: Process inline code (` ... `)
	codeRe := regexp.MustCompile("`([^`]+)`")
	text = codeRe.ReplaceAllString(text, "<code>$1</code>")
	
	// Step 7: Process bold (** ... **)
	boldRe := regexp.MustCompile(`\*\*(.+?)\*\*`)
	text = boldRe.ReplaceAllString(text, "<strong>$1</strong>")
	
	return text
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
