package paperscool

import "testing"

func TestResolveFeedURL(t *testing.T) {
	c := NewClient()

	cases := []struct {
		in   string
		want string
	}{
		{"cs.AI", "https://papers.cool/arxiv/cs.AI/feed"},
		{"arxiv/cs.CV", "https://papers.cool/arxiv/cs.CV/feed"},
		{"/arxiv/cs.LG/feed", "https://papers.cool/arxiv/cs.LG/feed"},
		{"https://papers.cool/arxiv/cs.CL/feed", "https://papers.cool/arxiv/cs.CL/feed"},
	}

	for _, tc := range cases {
		if got := c.resolveFeedURL(tc.in); got != tc.want {
			t.Fatalf("resolveFeedURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestExtractPaperID(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://papers.cool/arxiv/2602.22094", "2602.22094"},
		{"http://arxiv.org/abs/2602.22094v1", "2602.22094"},
		{"invalid-id", ""},
	}

	for _, tc := range cases {
		if got := extractPaperID(tc.in); got != tc.want {
			t.Fatalf("extractPaperID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestStripHTML(t *testing.T) {
	in := `<p><strong>Q1</strong>: 测试 &amp; 验证</p><div>内容&nbsp;A</div>`
	got := stripHTML(in)
	if got == "" || got == in {
		t.Fatalf("stripHTML should return cleaned text, got: %q", got)
	}
}
