# paper-radar

`paper-radar` 是一个用于论文追踪的 Go CLI：支持多数据源抓取、按关键词打分过滤、去重，并生成每日 Markdown 摘要。

## v0.3.0 功能

- 命令：`fetch`、`digest`、`run`
- 多数据源抓取：
  - `arxiv`（官方 arXiv API）
  - `paperscool`（papers.cool feed）
- `paperscool` 支持可选 Kimi 总结增强（`kimi_summary: true` 或 `-with-kimi`）
- YAML 配置支持：
  - `max_results`
  - `min_score`
  - `topics[].source`
  - `topics[].query`
  - `topics[].keywords`
  - `topics[].kimi_summary`
- 最低分过滤优先级：
  - CLI 覆盖 > topic `min_score` > 全局 `min_score` > 默认 `1`
- 本地状态与去重：`.paper-radar/state.json`
- 摘要输出：`outputs/YYYY-MM-DD.md`
- `run` 执行后可选飞书 Webhook 推送

## 快速开始

1) 复制示例配置：

```bash
cp config.example.yaml config.yaml
```

2) 构建：

```bash
go build ./cmd/paper-radar
```

## 命令说明

### 1) 抓取论文（fetch）

```bash
go run ./cmd/paper-radar fetch -config config.yaml
```

可选参数：

- `-state` 状态文件路径（默认 `.paper-radar/state.json`）
- `-max-results` 覆盖每个 topic 的最大抓取数
- `-min-score` 覆盖最低打分阈值
- `-with-kimi` 强制启用 papers.cool Kimi 总结增强

### 2) 生成摘要（digest）

```bash
go run ./cmd/paper-radar digest -state .paper-radar/state.json -out outputs
```

可选参数：

- `-date YYYY-MM-DD` 指定输出日期

### 3) 一键流程（run = fetch + digest + 可选通知）

```bash
go run ./cmd/paper-radar run -config config.yaml -out outputs
```

可选参数：

- `-state` 状态文件路径
- `-date YYYY-MM-DD`
- `-max-results`
- `-min-score`
- `-with-kimi`
- `-feishu-webhook https://open.feishu.cn/open-apis/bot/v2/hook/xxxxx`

## 配置示例

请参考 [`config.example.yaml`](./config.example.yaml)。

## 测试

```bash
go test ./...
```
