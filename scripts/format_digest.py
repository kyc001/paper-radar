#!/usr/bin/env python3
"""Reformat paper-radar digest: split single-line summaries into proper Markdown."""

import re
import sys
from pathlib import Path


def split_q_sections(summary: str) -> dict[str, str]:
    """Split summary into Q1-Q7 sections."""
    pattern = r'Q(\d+)\s*[:：]\s*'
    parts = re.split(pattern, summary)
    # parts = [before_Q1, '1', content1, '2', content2, ...]
    sections = {}
    i = 1
    while i < len(parts) - 1:
        qnum = parts[i]
        content = parts[i + 1].strip()
        sections[f'Q{qnum}'] = content
        i += 2
    return sections


def format_md_content(text: str) -> str:
    """Insert line breaks before markdown structural elements."""
    # Step 1: Handle headings - keep "## N. Title" or "### N. Title" together
    # Match heading patterns like "## 1. Title" or "### Title"
    text = re.sub(r'(?<!\n)\s+(#{2,4}\s+\d+\.\s)', r'\n\n\1', text)
    text = re.sub(r'(?<!\n)\s+(#{2,4}\s+[^\d#])', r'\n\n\1', text)

    # Step 2: Handle bold sub-headers like "**解决方案**：" or "**关键结果**："
    # These should start on a new line
    text = re.sub(r'(?<!\n)\s+(\*\*[^*]{2,60}\*\*\s*[:：])', r'\n\n\1', text)
    # After "**label**：", if followed by " - **", break into list
    text = re.sub(r'(\*\*\s*[:：])\s+(- )', r'\1\n\2', text)

    # Step 3: Handle standalone bold labels like "**早期探索与基线方法**"
    # that act as sub-section titles (not followed by : but followed by content)
    text = re.sub(r'(?<!\n)\s+(\*\*[^*]{2,40}\*\*)\s+(?=-\s)', r'\n\n\1\n', text)

    # Step 4: Handle list items - "- " at start of items
    # First handle "- **Bold**" pattern (common in paper summaries)
    text = re.sub(r'(?<=[^\n])\s+(- \*\*)', r'\n\1', text)
    # Then handle plain list items "- text" (not after newline already)
    text = re.sub(r'(?<=[^\n-])\s+(- [^*\n])', r'\n\1', text)
    # Handle list items after Chinese/word characters on same line
    text = re.sub(r'(?<=[\u4e00-\u9fff\w\)])\s+(- )', r'\n\1', text)

    # Step 5: Handle numbered list items "1. " etc
    text = re.sub(r'(?<!\n)\s+(\d+\.\s+(?!\*\*编码器|点云|可扩展|全局))', r'\n\n\1', text)

    # Step 6: Handle table rows - ensure each | row starts on new line
    text = re.sub(r'(?<!\n)\s*(\|[^|\n]+\|[^|\n]+\|)', r'\n\1', text)

    # Step 7: Handle $$ math blocks - own lines
    text = re.sub(r'(?<!\n)\s*(\$\$)', r'\n\n$$', text)
    text = re.sub(r'(\$\$)\s*(?!\n)', r'$$\n\n', text)

    # Step 8: Fix orphaned ### or ## (heading marker without content on same line)
    text = re.sub(r'\n(#{2,4})\n+(\d+\.)', r'\n\1 \2', text)
    text = re.sub(r'\n(#{2,4})\n+([^\n#])', r'\n\1 \2', text)

    # Step 9: Clean up excessive newlines (more than 2)
    text = re.sub(r'\n{3,}', '\n\n', text)

    # Step 10: Fix broken list items where "- " got separated from content
    text = re.sub(r'\n-\s*\n+\*\*', '\n- **', text)

    return text.strip()


def extract_q_title(content: str) -> tuple[str, str]:
    """Extract the first sentence as the Q title, return (title, rest)."""
    # The title is typically the first sentence before any heading or structure
    # Look for the first heading marker or significant structure
    match = re.match(r'^([^#\n|$]{10,200}?)(?:\s+(?=###|##|\*\*\d|1\.)|\s*$)', content)
    if match:
        title = match.group(1).strip()
        rest = content[match.end():].strip()
        return title, rest
    # Fallback: first ~100 chars
    idx = content.find('###')
    if idx == -1:
        idx = content.find('## ')
    if idx > 10:
        return content[:idx].strip(), content[idx:].strip()
    return '', content


def format_paper(title: str, score: str, topics: str, url: str, paper_id: str,
                 summary: str, published: str = '') -> str:
    """Format a single paper entry."""
    lines = []
    lines.append(f'## {title}')
    lines.append('')
    lines.append('| Field | Value |')
    lines.append('|-------|-------|')
    lines.append(f'| Score | {score} |')
    if topics:
        lines.append(f'| Topics | {topics} |')
    # Make URL clickable
    if url:
        arxiv_id = url.split('/')[-1] if '/' in url else url
        lines.append(f'| URL | [{arxiv_id}]({url}) |')
    if published:
        lines.append(f'| Published | {published} |')
    lines.append('')

    sections = split_q_sections(summary)

    q_titles = {
        'Q1': '这篇论文试图解决什么问题？',
        'Q2': '有哪些相关研究？',
        'Q3': '论文如何解决这个问题？',
        'Q4': '论文做了哪些实验？',
        'Q5': '有什么可以进一步探索的点？',
        'Q6': '总结一下论文的主要内容',
    }

    for qkey in ['Q1', 'Q2', 'Q3', 'Q4', 'Q5', 'Q6']:
        if qkey not in sections:
            continue
        content = sections[qkey]

        # Extract the Q title from content (it starts with the title text)
        default_title = q_titles.get(qkey, '')
        # Try to find the default title in the content and strip it
        if default_title and content.startswith(default_title):
            content = content[len(default_title):].strip()
        elif default_title:
            # Try fuzzy match - first sentence might be the title
            for t in [default_title]:
                idx = content.find(t)
                if idx >= 0 and idx < 5:
                    content = content[idx + len(t):].strip()
                    break

        lines.append(f'### {qkey}: {default_title}')
        lines.append('')
        formatted = format_md_content(content)
        lines.append(formatted)
        lines.append('')
        lines.append('')

    # Skip Q7 (Kimi promo)
    return '\n'.join(lines)


def parse_digest(text: str) -> list[dict]:
    """Parse the raw digest markdown into paper entries."""
    papers = []
    # Split by ## N. Title pattern
    paper_pattern = r'^## (\d+)\.\s+(.+)$'
    entries = re.split(r'(?=^## \d+\.)', text, flags=re.MULTILINE)

    for entry in entries:
        entry = entry.strip()
        if not entry:
            continue
        # Parse header
        header_match = re.match(paper_pattern, entry, re.MULTILINE)
        if not header_match:
            continue

        num = header_match.group(1)
        title = header_match.group(2).strip()
        rest = entry[header_match.end():].strip()

        # Parse metadata
        score = ''
        topics = ''
        url = ''
        paper_id = ''
        published = ''

        score_m = re.search(r'- Score:\s*(\d+)', rest)
        if score_m:
            score = score_m.group(1)
        topics_m = re.search(r'- Topics:\s*(.+)', rest)
        if topics_m:
            topics = topics_m.group(1).strip()
        url_m = re.search(r'- URL:\s*(\S+)', rest)
        if url_m:
            url = url_m.group(1).strip()
        id_m = re.search(r'- ID:\s*`?(\S+?)`?$', rest, re.MULTILINE)
        if id_m:
            paper_id = id_m.group(1).strip()
        pub_m = re.search(r'- Published:\s*(\S+)', rest)
        if pub_m:
            published = pub_m.group(1).strip()

        # Extract summary (everything after metadata block)
        # Find where Q1 starts
        q1_idx = rest.find('Q1')
        if q1_idx >= 0:
            summary = rest[q1_idx:]
        else:
            summary = ''

        papers.append({
            'num': num,
            'title': title,
            'score': score,
            'topics': topics,
            'url': url,
            'id': paper_id,
            'published': published,
            'summary': summary,
        })

    return papers


def main():
    if len(sys.argv) < 2:
        # Default: today or latest
        outputs_dir = Path(__file__).parent.parent / 'outputs'
        md_files = sorted(outputs_dir.glob('????-??-??.md'))
        if not md_files:
            print("No digest files found in outputs/")
            sys.exit(1)
        input_path = md_files[-1]
    else:
        input_path = Path(sys.argv[1])

    if not input_path.exists():
        print(f"File not found: {input_path}")
        sys.exit(1)

    text = input_path.read_text(encoding='utf-8')

    # Extract date from title
    date_match = re.search(r'# Paper Radar Digest (\d{4}-\d{2}-\d{2})', text)
    date_str = date_match.group(1) if date_match else 'Unknown'

    papers = parse_digest(text)
    if not papers:
        print("No papers found in digest")
        sys.exit(1)

    # Build output
    out_lines = [f'# Paper Radar Digest {date_str}', '']
    out_lines.append(f'> {len(papers)} papers | Auto-formatted')
    out_lines.append('')

    for i, paper in enumerate(papers):
        numbered_title = f"{paper['num']}. {paper['title']}"
        formatted = format_paper(
            numbered_title,
            paper['score'],
            paper['topics'],
            paper['url'],
            paper['id'],
            paper['summary'],
            paper['published'],
        )
        out_lines.append(formatted)
        if i < len(papers) - 1:
            out_lines.append('---')
            out_lines.append('')

    output = '\n'.join(out_lines)

    # Write output
    output_path = input_path.with_name(input_path.stem + '-formatted.md')
    output_path.write_text(output, encoding='utf-8')
    print(f"Formatted {len(papers)} papers -> {output_path}")


if __name__ == '__main__':
    main()
