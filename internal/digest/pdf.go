package digest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/kyc001/paper-radar/internal/model"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// WritePDF generates a PDF by converting the markdown digest to HTML and
// rendering it via headless Chrome. KaTeX is loaded from CDN to render
// LaTeX math formulas.
func WritePDF(outputDir string, date time.Time, papers []model.ScoredPaper) (string, error) {
	filename := date.Format("2006-01-02") + ".pdf"
	path := filepath.Join(outputDir, filename)

	md := BuildMarkdown(date, papers)

	// Markdown → HTML body
	converter := goldmark.New(goldmark.WithExtensions(extension.Table))
	var buf bytes.Buffer
	if err := converter.Convert([]byte(md), &buf); err != nil {
		return "", fmt.Errorf("markdown to html: %w", err)
	}

	html := wrapHTML(buf.String())

	// Write HTML to temp file so Chrome can load external resources (KaTeX CDN)
	tmpFile, err := os.CreateTemp("", "paper-radar-*.html")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(html); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	// HTML → PDF via headless Chrome
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var pdfBuf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate("file://"+tmpPath),
		// Wait for KaTeX to finish rendering
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			if err != nil {
				return fmt.Errorf("print to PDF: %w", err)
			}
			pdfBuf = buf
			return nil
		}),
	); err != nil {
		return "", fmt.Errorf("chromedp: %w", err)
	}

	if err := os.WriteFile(path, pdfBuf, 0644); err != nil {
		return "", fmt.Errorf("write PDF file: %w", err)
	}
	return path, nil
}

func wrapHTML(body string) string {
	return `<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="UTF-8">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.16.21/dist/katex.min.css">
<script src="https://cdn.jsdelivr.net/npm/katex@0.16.21/dist/katex.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/katex@0.16.21/dist/contrib/auto-render.min.js"></script>
<style>
@page { size: A4; margin: 20mm 16mm; }
* { box-sizing: border-box; }
body {
  font-family: "Noto Sans CJK SC", "Noto Sans SC", "Microsoft YaHei", "PingFang SC",
               -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-size: 11pt; line-height: 1.8; color: #1f2937;
  margin: 0; padding: 0;
}
h1 { font-size: 22pt; color: #2563eb; border-bottom: 3px solid #2563eb; padding-bottom: 8px; }
h2 { font-size: 16pt; color: #1e40af; margin-top: 32px; page-break-before: always; }
h2:first-of-type { page-break-before: avoid; }
h3 { font-size: 13pt; color: #1d4ed8; margin-top: 20px; }
blockquote {
  border-left: 4px solid #93c5fd; background: #eff6ff;
  margin: 12px 0; padding: 8px 16px; color: #374151;
}
table { width: 100%; border-collapse: collapse; font-size: 10pt; margin: 12px 0; }
th, td { border: 1px solid #d1d5db; padding: 8px 12px; text-align: left; }
th { background: #f1f5f9; font-weight: 700; }
tr:nth-child(even) td { background: #f8fafc; }
ul, ol { padding-left: 24px; }
li { margin: 4px 0; }
code { background: #f3f4f6; padding: 2px 6px; border-radius: 3px; font-size: 10pt; }
strong { color: #111827; }
hr { border: none; border-top: 2px dashed #d1d5db; margin: 32px 0; page-break-after: avoid; }
a { color: #2563eb; text-decoration: none; }
@media print {
  h2 { page-break-before: always; }
  h2:first-of-type { page-break-before: avoid; }
  table, blockquote { page-break-inside: avoid; }
}
</style>
</head><body>` + body + `
<script>
document.addEventListener("DOMContentLoaded", function() {
  renderMathInElement(document.body, {
    delimiters: [
      {left: "$$", right: "$$", display: true},
      {left: "$", right: "$", display: false}
    ],
    throwOnError: false
  });
});
</script>
</body></html>`
}
