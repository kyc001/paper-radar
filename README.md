# paper-radar

`paper-radar` 是一个用于论文追踪的 Go CLI：支持多数据源抓取、按关键词打分过滤、去重，并生成每日 Markdown/HTML/PDF 摘要。

## 架构概览

```
┌─────────────────────────────────────────────────────────┐
│                     CLI (cmd/paper-radar)                │
│              fetch  |  digest  |  run (全流程)           │
└───────┬─────────────┬──────────────┬────────────────────┘
        │             │              │
   ┌────▼────┐   ┌────▼────┐   ┌────▼────┐
   │  Fetch  │   │ Digest  │   │ Notify  │
   │ (数据源) │   │ (格式化) │   │ (推送)  │
   └────┬────┘   └────┬────┘   └────┬────┘
        │             │              │
   ┌────▼────────┐ ┌──▼───────┐  ┌──▼──────────┐
   │ arxiv API   │ │ Markdown │  │ Feishu Bot  │
   │ papers.cool │ │ HTML     │  │ (Webhook)   │
   │ + Kimi 总结 │ │ PDF      │  └─────────────┘
   └────┬────────┘ └──────────┘
        │
   ┌────▼──────────────────────┐
   │  state.json (去重+pending) │
   └───────────────────────────┘
```

### 数据流

```
arXiv API / papers.cool RSS
    → 论文元数据 (Title, Abstract, URL, PublishedAt)
    → 关键词打分 & 去重
    → state.json (pending papers)

papers.cool Kimi API (可选)
    → HTML 格式的 Q&A 摘要 (<div class="faq-a"> 包裹 Markdown)
    → htmlToMarkdown() 提取并保留格式
    → Paper.Summary (含完整换行和 Markdown 结构)

Digest 生成
    → splitQSections() 拆分 Q1-Q6 段落 (跳过 Q7 Kimi 推广)
    → BuildMarkdown() → YYYY-MM-DD.md (结构化 Markdown)
    → WriteHTML() → YYYY-MM-DD.html (带样式的 HTML)
    → WritePDF() → YYYY-MM-DD.pdf (可选)
```

### 关键模块

| 文件 | 功能 |
|------|------|
| `cmd/paper-radar/main.go` | CLI 入口，解析命令和参数 |
| `internal/arxiv/client.go` | arXiv 官方 API 抓取 |
| `internal/paperscool/client.go` | papers.cool RSS 抓取 + Kimi 摘要获取 |
| `internal/scoring/scorer.go` | 关键词匹配打分 |
| `internal/state/state.go` | 本地状态管理与去重 |
| `internal/digest/markdown.go` | Markdown 摘要生成 (Q 段落拆分、元数据表格) |
| `internal/digest/html.go` | HTML 摘要生成 (带 CSS 样式) |
| `internal/digest/pdf.go` | PDF 导出 |
| `internal/notify/feishu.go` | 飞书 Webhook 推送 (自动分片) |
| `internal/model/model.go` | 核心数据结构 (Paper, ScoredPaper) |

## 功能特性

- **多数据源**：`arxiv` (官方 API) + `paperscool` (papers.cool feed)
- **Kimi 摘要增强**：papers.cool 集成 Kimi 论文总结，自动生成 Q1-Q6 结构化摘要
- **智能格式化**：`htmlToMarkdown()` 保留 Kimi 返回的完整 Markdown 结构（标题、列表、表格、公式块）
- **关键词打分**：YAML 配置关键词列表，按匹配数量和位置打分
- **去重机制**：基于 arXiv ID 的本地状态去重，跨次运行不重复推送
- **多格式输出**：Markdown + HTML + PDF
- **飞书推送**：长消息自动分片，适配飞书消息长度限制
- **摘要条数控制**：`-top N` 只输出前 N 条，剩余保留在 pending

## 快速开始

1) 复制示例配置：

```bash
cp config.example.yaml config.yaml
```

2) 构建：

```bash
go build -o paper-radar ./cmd/paper-radar
```

3) 一键运行：

```bash
./paper-radar run -config configs/focus-3d-video-kimi.yaml -out outputs -with-kimi
```

## 命令说明

### 1) 抓取论文（fetch）

```bash
./paper-radar fetch -config config.yaml
```

可选参数：

- `-state` 状态文件路径（默认 `.paper-radar/state.json`）
- `-max-results` 覆盖每个 topic 的最大抓取数
- `-min-score` 覆盖最低打分阈值
- `-with-kimi` 强制启用 papers.cool Kimi 总结增强

### 2) 生成摘要（digest）

```bash
./paper-radar digest -state .paper-radar/state.json -out outputs
```

可选参数：

- `-date YYYY-MM-DD` 指定输出日期
- `-top 20` 仅输出前 20 篇（其余保留在 pending，留待下次 digest）
- `-html` 同时生成 HTML 版本
- `-pdf` 同时生成 PDF 版本

### 3) 一键流程（run = fetch + digest + 可选通知）

```bash
./paper-radar run -config config.yaml -out outputs
```

可选参数：

- `-state` 状态文件路径
- `-date YYYY-MM-DD`
- `-max-results`
- `-min-score`
- `-top`
- `-with-kimi`
- `-feishu-webhook https://open.feishu.cn/open-apis/bot/v2/hook/xxxxx`
- `-notify-max-chars 2800`（飞书单条消息最大字符数，超出自动分片）
- `-html` 同时生成 HTML 版本
- `-pdf` 同时生成 PDF 版本

## YAML 配置

```yaml
max_results: 50          # 每个 topic 最大抓取数
min_score: 1             # 全局最低分阈值
feishu_webhook: "..."    # 飞书 Webhook 地址

topics:
  - source: paperscool         # 数据源: arxiv / paperscool
    query: cs.CV               # arXiv 分类或 papers.cool 频道
    kimi_summary: true         # 启用 Kimi 摘要 (仅 paperscool)
    min_score: 5               # topic 级别最低分
    keywords:
      - {word: "3D", weight: 10}
      - {word: "video generation", weight: 8}
      - {word: "diffusion", weight: 5}
```

最低分过滤优先级：CLI `-min-score` > topic `min_score` > 全局 `min_score` > 默认 `1`

## 摘要输出格式

每篇论文的 Markdown 摘要结构：

```markdown
## N. Paper Title

| Field | Value |
|-------|-------|
| Score | 101 |
| Topics | 3D/Video Training-Free |
| URL | [2602.23153](https://papers.cool/arxiv/2602.23153) |
| Published | 2026-02-27 |

### Q1: 这篇论文试图解决什么问题？

(结构化内容：段落、列表、表格、公式)

### Q2: 有哪些相关研究？
...
### Q6: 总结一下论文的主要内容
...
---
```

## 研究方向预设

内置示例：[`configs/focus-3d-video-kimi.yaml`](./configs/focus-3d-video-kimi.yaml)

```bash
./paper-radar run \
  -config configs/focus-3d-video-kimi.yaml \
  -with-kimi \
  -top 5 \
  -out outputs
```

## 测试

```bash
go test ./...
```

## 本地状态

- 状态文件：`.paper-radar/state.json`
- 包含 `seen`（已处理的 arXiv ID 集合）和 `pending`（待生成摘要的论文）
- `fetch` 写入 pending，`digest` 消费 pending 并标记 seen
- 支持跨次运行去重
