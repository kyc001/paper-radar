# paper-radar

`paper-radar` is a Go CLI for fetching arXiv papers, deduplicating by paper ID, scoring by keyword frequency, and generating a daily Markdown digest.

## v0.2.0 Features

- Go module: `github.com/kyc001/paper-radar`
- CLI commands: `fetch`, `digest`, `run`
- YAML config with global/per-topic controls:
  - `max_results`
  - `min_score`
  - `topics[].keywords` / `topics[].query`
- arXiv Atom API fetch via `net/http` + `encoding/xml`
- Local dedupe/state in `.paper-radar/state.json`
- Keyword frequency scoring on title + summary
- Minimum score filtering with precedence:
  - CLI override > topic `min_score` > global `min_score` > default `1`
- Markdown digest output to `outputs/YYYY-MM-DD.md`

## Setup

1. Copy sample config:

```bash
cp config.example.yaml config.yaml
```

2. Build:

```bash
go build ./cmd/paper-radar
```

## Commands

### Fetch papers

```bash
go run ./cmd/paper-radar fetch -config config.yaml
```

Optional flags:

- `-state` state file path (default `.paper-radar/state.json`)
- `-max-results` override max results per topic
- `-min-score` override minimum score threshold

### Generate digest

```bash
go run ./cmd/paper-radar digest -state .paper-radar/state.json -out outputs
```

Optional flags:

- `-date YYYY-MM-DD` write digest for a specific date

### One-shot pipeline (fetch + digest)

```bash
go run ./cmd/paper-radar run -config config.yaml -out outputs
```

Optional flags:

- `-state` state file path
- `-date YYYY-MM-DD`
- `-max-results`
- `-min-score`

## Testing

Run all tests:

```bash
go test ./...
```
