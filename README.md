# paper-radar

`paper-radar` is a Go CLI for fetching papers, deduplicating by paper ID, scoring by keyword frequency, and generating a daily Markdown digest.

## v0.3.0 Features

- CLI commands: `fetch`, `digest`, `run`
- Multi-source collection:
  - `arxiv` (official arXiv API)
  - `paperscool` (papers.cool feed)
- `paperscool` optional Kimi summary enrichment (`kimi_summary: true` or `-with-kimi`)
- YAML config with global/per-topic controls:
  - `max_results`
  - `min_score`
  - `topics[].source`
  - `topics[].query`
  - `topics[].keywords`
  - `topics[].kimi_summary`
- Minimum score filtering precedence:
  - CLI override > topic `min_score` > global `min_score` > default `1`
- Local dedupe/state in `.paper-radar/state.json`
- Markdown digest output to `outputs/YYYY-MM-DD.md`
- Optional Feishu webhook notification after `run`

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
- `-with-kimi` force papers.cool Kimi summary enrichment

### Generate digest

```bash
go run ./cmd/paper-radar digest -state .paper-radar/state.json -out outputs
```

Optional flags:

- `-date YYYY-MM-DD` write digest for a specific date

### One-shot pipeline (fetch + digest + optional Feishu notify)

```bash
go run ./cmd/paper-radar run -config config.yaml -out outputs
```

Optional flags:

- `-state` state file path
- `-date YYYY-MM-DD`
- `-max-results`
- `-min-score`
- `-with-kimi`
- `-feishu-webhook https://open.feishu.cn/open-apis/bot/v2/hook/xxxxx`

## Config Example

See [`config.example.yaml`](./config.example.yaml).

## Testing

```bash
go test ./...
```
