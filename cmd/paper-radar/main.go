package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kyc001/paper-radar/internal/app"
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
	fs.Parse(args)

	result, err := app.RunFetch(ctx, app.FetchOptions{
		ConfigPath: *configPath,
		StatePath:  *statePath,
		MaxResults: *maxResults,
		MinScore:   *minScore,
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
	fs.Parse(args)

	date := parseDateOrNow(*dateStr)

	path, count, err := app.RunDigest(app.DigestOptions{
		StatePath: *statePath,
		OutputDir: *outputDir,
		Date:      date,
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
	fs.Parse(args)

	fetchResult, err := app.RunFetch(ctx, app.FetchOptions{
		ConfigPath: *configPath,
		StatePath:  *statePath,
		MaxResults: *maxResults,
		MinScore:   *minScore,
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
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "run failed in digest stage: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("digest: path=%s papers=%d\n", path, count)
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
	fmt.Fprintln(os.Stderr, "  paper-radar fetch  -config config.yaml")
	fmt.Fprintln(os.Stderr, "  paper-radar digest -out outputs")
	fmt.Fprintln(os.Stderr, "  paper-radar run    -config config.yaml -out outputs")
}
