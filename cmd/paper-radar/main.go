package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kyc001/paper-radar/internal/app"
	"github.com/kyc001/paper-radar/internal/config"
	"github.com/kyc001/paper-radar/internal/notify"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "fetch":
		runFetch(ctx, os.Args[2:])
	case "digest":
		runDigest(os.Args[2:])
	case "run":
		runAll(ctx, os.Args[2:])
	default:
		printUsage()
		os.Exit(2)
	}
}

func runFetch(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to YAML config file")
	statePath := fs.String("state", app.DefaultStatePath, "Path to JSON state file")
	maxResults := fs.Int("max-results", 0, "Override max results per topic")
	minScore := fs.Int("min-score", 1, "Override minimum score threshold")
	withKimi := fs.Bool("with-kimi", false, "Enable papers.cool Kimi summary enrichment")
	fs.Parse(args)

	result, err := app.RunFetch(ctx, app.FetchOptions{
		ConfigPath: *configPath,
		StatePath:  *statePath,
		MaxResults: *maxResults,
		MinScore:   *minScore,
		WithKimi:   *withKimi,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("fetched=%d queued=%d topics=%d\n", result.Fetched, result.Queued, result.Topics)
}

func runDigest(args []string) {
	fs := flag.NewFlagSet("digest", flag.ExitOnError)
	statePath := fs.String("state", app.DefaultStatePath, "Path to JSON state file")
	outputDir := fs.String("out", "outputs", "Output directory for markdown digest")
	dateStr := fs.String("date", "", "Digest date (YYYY-MM-DD), defaults to today")
	topN := fs.Int("top", 0, "Only emit top N papers in this digest (0 means all)")
	asHTML := fs.Bool("html", false, "Generate HTML output (can be printed to PDF via browser)")
	fs.Parse(args)

	date := parseDateOrNow(*dateStr)

	path, count, err := app.RunDigest(app.DigestOptions{
		StatePath: *statePath,
		OutputDir: *outputDir,
		Date:      date,
		TopN:      *topN,
		AsHTML:    *asHTML,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "digest failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("digest=%s papers=%d\n", path, count)
}

func runAll(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("config", "config.yaml", "Path to YAML config file")
	statePath := fs.String("state", app.DefaultStatePath, "Path to JSON state file")
	outputDir := fs.String("out", "outputs", "Output directory for markdown digest")
	dateStr := fs.String("date", "", "Digest date (YYYY-MM-DD), defaults to today")
	maxResults := fs.Int("max-results", 0, "Override max results per topic")
	minScore := fs.Int("min-score", 1, "Override minimum score threshold")
	topN := fs.Int("top", 0, "Only emit top N papers in this digest (0 means all)")
	withKimi := fs.Bool("with-kimi", false, "Enable papers.cool Kimi summary enrichment")
	feishuWebhook := fs.String("feishu-webhook", "", "Feishu bot webhook URL for digest notification")
	notifyMaxChars := fs.Int("notify-max-chars", 2800, "Max characters per Feishu message chunk")
	asHTML := fs.Bool("html", false, "Generate HTML output (can be printed to PDF via browser)")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run failed loading config: %v\n", err)
		os.Exit(1)
	}
	resolvedWebhook := resolveWebhook(*feishuWebhook, cfg.FeishuWebhook)

	fetchResult, err := app.RunFetch(ctx, app.FetchOptions{
		ConfigPath: *configPath,
		StatePath:  *statePath,
		MaxResults: *maxResults,
		MinScore:   *minScore,
		WithKimi:   *withKimi,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "run failed in fetch stage: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("fetch: fetched=%d queued=%d topics=%d\n", fetchResult.Fetched, fetchResult.Queued, fetchResult.Topics)

	date := parseDateOrNow(*dateStr)
	path, count, err := app.RunDigest(app.DigestOptions{
		StatePath: *statePath,
		OutputDir: *outputDir,
		Date:      date,
		TopN:      *topN,
		AsHTML:    *asHTML,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "run failed in digest stage: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("digest: path=%s papers=%d\n", path, count)

	if resolvedWebhook != "" {
		content := ""
		if raw, readErr := os.ReadFile(path); readErr == nil {
			content = strings.TrimSpace(string(raw))
		}
		text := fmt.Sprintf("paper-radar run completed\nfetch: fetched=%d queued=%d topics=%d\ndigest: papers=%d\nfile: %s", fetchResult.Fetched, fetchResult.Queued, fetchResult.Topics, count, path)
		if content != "" {
			text += "\n\n" + content
		}
		if err := notify.NewFeishuWebhook().SendLongText(ctx, resolvedWebhook, text, *notifyMaxChars); err != nil {
			fmt.Fprintf(os.Stderr, "feishu notify failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("feishu: notification sent")
	}
}

func resolveWebhook(cliValue, configValue string) string {
	candidates := []string{
		strings.TrimSpace(cliValue),
		strings.TrimSpace(configValue),
		strings.TrimSpace(os.Getenv("PAPER_RADAR_FEISHU_WEBHOOK")),
	}
	for _, c := range candidates {
		if c != "" {
			return c
		}
	}
	return ""
}

func parseDateOrNow(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid date %q: %v\n", dateStr, err)
		os.Exit(2)
	}
	return parsed
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "paper-radar: track and score arXiv papers")
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  paper-radar fetch  -config config.yaml [-with-kimi]")
	fmt.Fprintln(os.Stderr, "  paper-radar digest -out outputs [-top 20] [-html]")
	fmt.Fprintln(os.Stderr, "  paper-radar run    -config config.yaml -out outputs [-top 20] [-with-kimi] [-feishu-webhook URL] [-notify-max-chars 2800] [-html]")
}
