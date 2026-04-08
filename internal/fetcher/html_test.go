package fetcher

import (
	"strings"
	"testing"
)

func TestHTMLToText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple paragraph",
			input:    "<p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "Nested tags",
			input:    "<p>Hello <b>World</b></p>",
			expected: "Hello **World**",
		},
		{
			name:     "Headers",
			input:    "<h1>Title</h1><p>Body</p>",
			expected: "# Title\n\nBody",
		},
		{
			name:     "Lists",
			input:    "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expected: "- Item 1\n- Item 2",
		},
		{
			name:     "Code block",
			input:    "<pre><code>fmt.Println(\"Hello\")</code></pre>",
			expected: "```\nfmt.Println(\"Hello\")\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := htmlToText(tt.input)
			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
