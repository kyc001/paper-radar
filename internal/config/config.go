package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MaxResults int
	MinScore   int
	Topics     []Topic
}

type Topic struct {
	Name        string
	Source      string
	Query       string
	Keywords    []string
	MaxResults  int
	MinScore    int
	KimiSummary bool
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg, err := parseYAMLSubset(string(data))
	if err != nil {
		return Config{}, err
	}

	if err := (&cfg).Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Topics) == 0 {
		return fmt.Errorf("config must include at least one topic")
	}

	if c.MinScore < 0 {
		return fmt.Errorf("min_score must be >= 0")
	}

	for i, topic := range c.Topics {
		topic.Name = strings.TrimSpace(topic.Name)
		if topic.Name == "" {
			return fmt.Errorf("topic[%d] must have a name", i)
		}

		topic.Source = strings.ToLower(strings.TrimSpace(topic.Source))
		if topic.Source == "" {
			topic.Source = "arxiv"
		}
		if topic.Source != "arxiv" && topic.Source != "paperscool" {
			return fmt.Errorf("topic[%d] (%s) source must be arxiv or paperscool", i, topic.Name)
		}

		topic.Keywords = normalizeKeywords(topic.Keywords)
		topic.Query = strings.TrimSpace(topic.Query)

		if len(topic.Keywords) == 0 {
			return fmt.Errorf("topic[%d] (%s) must have at least one keyword", i, topic.Name)
		}
		if topic.MinScore < 0 {
			return fmt.Errorf("topic[%d] (%s) min_score must be >= 0", i, topic.Name)
		}

		c.Topics[i] = topic
	}

	return nil
}

func (c Config) TopicQuery(topic Topic) string {
	if topic.Query != "" {
		return topic.Query
	}

	if topic.Source == "paperscool" {
		// default papers.cool category when query is omitted
		return "cs.AI"
	}

	parts := make([]string, 0, len(topic.Keywords))
	for _, keyword := range topic.Keywords {
		parts = append(parts, fmt.Sprintf("all:\"%s\"", keyword))
	}

	return strings.Join(parts, " OR ")
}

func (c Config) EffectiveMaxResults(topic Topic, override int) int {
	if override > 0 {
		return override
	}
	if topic.MaxResults > 0 {
		return topic.MaxResults
	}
	if c.MaxResults > 0 {
		return c.MaxResults
	}
	return 25
}

func (c Config) EffectiveMinScore(topic Topic, override int) int {
	if override > 0 {
		return override
	}
	if topic.MinScore > 0 {
		return topic.MinScore
	}
	if c.MinScore > 0 {
		return c.MinScore
	}
	return 1
}

func normalizeKeywords(keywords []string) []string {
	normalized := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}
		normalized = append(normalized, keyword)
	}
	return normalized
}

func parseYAMLSubset(content string) (Config, error) {
	var cfg Config
	var current *Topic
	inTopics := false
	inKeywords := false

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0

	flushTopic := func() {
		if current != nil {
			cfg.Topics = append(cfg.Topics, *current)
			current = nil
		}
	}

	for scanner.Scan() {
		lineNo++
		raw := strings.TrimRight(scanner.Text(), " \t")
		line := strings.TrimSpace(stripInlineComment(raw))
		if line == "" {
			continue
		}

		indent := len(raw) - len(strings.TrimLeft(raw, " "))

		if strings.HasPrefix(line, "- ") {
			item := strings.TrimSpace(strings.TrimPrefix(line, "- "))
			if inKeywords && !strings.HasPrefix(item, "name:") {
				if current == nil {
					return Config{}, fmt.Errorf("line %d: keyword item appears before a topic", lineNo)
				}
				current.Keywords = append(current.Keywords, parseScalar(item))
				continue
			}

			if !inTopics {
				return Config{}, fmt.Errorf("line %d: list item appears outside topics", lineNo)
			}

			flushTopic()
			current = &Topic{}

			if strings.HasPrefix(item, "name:") {
				current.Name = parseScalar(strings.TrimSpace(strings.TrimPrefix(item, "name:")))
			} else {
				return Config{}, fmt.Errorf("line %d: topic item must start with '- name:'", lineNo)
			}

			inKeywords = false
			continue
		}

		if strings.HasPrefix(line, "topics:") {
			inTopics = true
			inKeywords = false
			continue
		}

		if strings.HasPrefix(line, "max_results:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "max_results:"))
			n, err := strconv.Atoi(value)
			if err != nil {
				return Config{}, fmt.Errorf("line %d: invalid max_results %q", lineNo, value)
			}
			if current != nil && indent > 0 {
				current.MaxResults = n
			} else {
				cfg.MaxResults = n
			}
			inKeywords = false
			continue
		}

		if strings.HasPrefix(line, "min_score:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "min_score:"))
			n, err := strconv.Atoi(value)
			if err != nil {
				return Config{}, fmt.Errorf("line %d: invalid min_score %q", lineNo, value)
			}
			if current != nil && indent > 0 {
				current.MinScore = n
			} else {
				cfg.MinScore = n
			}
			inKeywords = false
			continue
		}

		if strings.HasPrefix(line, "name:") {
			if current == nil {
				return Config{}, fmt.Errorf("line %d: name appears outside a topic", lineNo)
			}
			current.Name = parseScalar(strings.TrimSpace(strings.TrimPrefix(line, "name:")))
			inKeywords = false
			continue
		}

		if strings.HasPrefix(line, "source:") {
			if current == nil {
				return Config{}, fmt.Errorf("line %d: source appears outside a topic", lineNo)
			}
			current.Source = parseScalar(strings.TrimSpace(strings.TrimPrefix(line, "source:")))
			inKeywords = false
			continue
		}

		if strings.HasPrefix(line, "query:") {
			if current == nil {
				return Config{}, fmt.Errorf("line %d: query appears outside a topic", lineNo)
			}
			current.Query = parseScalar(strings.TrimSpace(strings.TrimPrefix(line, "query:")))
			inKeywords = false
			continue
		}

		if strings.HasPrefix(line, "keywords:") {
			if current == nil {
				return Config{}, fmt.Errorf("line %d: keywords appears outside a topic", lineNo)
			}
			inKeywords = true
			continue
		}

		if strings.HasPrefix(line, "kimi_summary:") {
			if current == nil {
				return Config{}, fmt.Errorf("line %d: kimi_summary appears outside a topic", lineNo)
			}
			value := parseScalar(strings.TrimSpace(strings.TrimPrefix(line, "kimi_summary:")))
			b, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return Config{}, fmt.Errorf("line %d: invalid kimi_summary %q", lineNo, value)
			}
			current.KimiSummary = b
			inKeywords = false
			continue
		}

		return Config{}, fmt.Errorf("line %d: unsupported config line %q", lineNo, line)
	}

	if err := scanner.Err(); err != nil {
		return Config{}, err
	}

	flushTopic()
	return cfg, nil
}

func parseScalar(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func stripInlineComment(line string) string {
	inSingle := false
	inDouble := false

	for i, ch := range line {
		switch ch {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return strings.TrimSpace(line[:i])
			}
		}
	}

	return line
}
