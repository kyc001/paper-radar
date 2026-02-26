package notify

import "testing"

func TestSplitTextRespectsLimit(t *testing.T) {
	text := "第一段\n第二段\n第三段\n第四段\n第五段"
	parts := splitText(text, 8)
	if len(parts) < 2 {
		t.Fatalf("expected multiple parts, got %d", len(parts))
	}
	for _, p := range parts {
		if len([]rune(p)) > 8 {
			t.Fatalf("chunk too long: %q", p)
		}
	}
}

func TestSplitTextKeepsSingleChunk(t *testing.T) {
	text := "hello world"
	parts := splitText(text, 50)
	if len(parts) != 1 || parts[0] != text {
		t.Fatalf("unexpected split result: %#v", parts)
	}
}
