package planner

import (
	"github.com/jluntpcty/contextual/internal/types"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Jira: CTX-123", "jira-ctx-123"},
		{"Some / weird & characters", "some-weird-characters"},
		{"   leading and trailing   ", "leading-and-trailing"},
		{"This is a very long title that should definitely be truncated because it exceeds the eighty character limit that we have set for our filesystem slugs in this system so let's make it really long", "this-is-a-very-long-title-that-should-definitely-be-truncated-because-it-exceeds"},
	}

	for _, test := range tests {
		result := slugify(test.input)
		if result != test.expected {
			t.Errorf("slugify(%q) = %q; want %q", test.input, test.expected, result)
		}
	}
}

func TestItemSlug(t *testing.T) {
	tests := []struct {
		item     types.Item
		expected string
	}{
		{
			item:     types.Item{Type: types.ItemTypeJira, ID: "CTX-123"},
			expected: "CTX-123",
		},
		{
			item:     types.Item{Type: types.ItemTypeConfluence, ID: "456", Title: "My Page"},
			expected: "my-page",
		},
		{
			item:     types.Item{Type: types.ItemTypeConfluence, ID: "456"},
			expected: "confluence-456",
		},
		{
			item:     types.Item{Type: types.ItemTypeWeb, URL: "https://example.com/foo", Title: "Example Page"},
			expected: "example-page",
		},
		{
			item:     types.Item{Type: types.ItemTypeWeb, URL: "https://example.com/foo"},
			expected: "example-com-foo",
		},
	}

	for _, test := range tests {
		result := ItemSlug(test.item)
		if result != test.expected {
			t.Errorf("ItemSlug(%+v) = %q; want %q", test.item, test.expected, result)
		}
	}
}
