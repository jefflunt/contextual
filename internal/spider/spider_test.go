package spider

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jluntpcty/contextual/internal/config"
	"github.com/jluntpcty/contextual/internal/types"
)

func newTestSpider(host string) *Spider {
	return New(&config.Config{Atlassian: config.AtlassianConfig{Host: host, APIUser: "test@example.com", APIToken: "token"}}, nil)
}

func TestParseItemJiraKey(t *testing.T) {
	s := newTestSpider("example.atlassian.net")
	item, err := s.ParseItem("CTX-1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != types.ItemTypeJira {
		t.Errorf("expected jira, got %s", item.Type)
	}
	if item.ID != "CTX-1234" {
		t.Errorf("expected CTX-1234, got %s", item.ID)
	}
	if !strings.Contains(item.URL, "browse/CTX-1234") {
		t.Errorf("expected browse URL, got %s", item.URL)
	}
}

func TestParseItemConfluenceNumericID(t *testing.T) {
	s := newTestSpider("example.atlassian.net")
	item, err := s.ParseItem("12345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != types.ItemTypeConfluence {
		t.Errorf("expected confluence, got %s", item.Type)
	}
	if item.ID != "12345678" {
		t.Errorf("expected 12345678, got %s", item.ID)
	}
}

func TestParseItemShortNumericNotConfluence(t *testing.T) {
	s := newTestSpider("example.atlassian.net")
	// Fewer than 8 digits → not recognised as Confluence ID, not a valid arg.
	_, err := s.ParseItem("1234")
	if err == nil {
		t.Fatal("expected error for short numeric string")
	}
}

func TestParseItemJiraURL(t *testing.T) {
	s := newTestSpider("example.atlassian.net")
	rawURL := "https://example.atlassian.net/browse/PROJ-42"
	item, err := s.ParseItem(rawURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != types.ItemTypeJira {
		t.Errorf("expected jira, got %s", item.Type)
	}
	if item.ID != "PROJ-42" {
		t.Errorf("expected PROJ-42, got %s", item.ID)
	}
}

func TestParseItemConfluenceURL(t *testing.T) {
	s := newTestSpider("example.atlassian.net")
	rawURL := "https://example.atlassian.net/wiki/spaces/ENG/pages/123456789/My+Page"
	item, err := s.ParseItem(rawURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != types.ItemTypeConfluence {
		t.Errorf("expected confluence, got %s", item.Type)
	}
	if item.ID != "123456789" {
		t.Errorf("expected 123456789, got %s", item.ID)
	}
}

func TestParseItemGenericWebURL(t *testing.T) {
	s := newTestSpider("example.atlassian.net")
	rawURL := "https://github.com/some/repo"
	item, err := s.ParseItem(rawURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != types.ItemTypeWeb {
		t.Errorf("expected web, got %s", item.Type)
	}
	if item.URL != rawURL {
		t.Errorf("expected %s, got %s", rawURL, item.URL)
	}
}

func TestParseItemUnrecognised(t *testing.T) {
	s := newTestSpider("example.atlassian.net")
	_, err := s.ParseItem("not-valid")
	if err == nil {
		t.Fatal("expected error for unrecognised argument")
	}
}

func TestRunWebItem(t *testing.T) {
	// Mock HTTP server returning a simple HTML page.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body><h1>Hello</h1></body></html>`))
	}))
	defer srv.Close()

	s := newTestSpider("")
	items, err := s.Run([]string{srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != types.ItemTypeWeb {
		t.Errorf("expected web item, got %s", items[0].Type)
	}
	if !strings.Contains(items[0].Content, "Hello") {
		t.Errorf("expected content to contain 'Hello', got %q", items[0].Content)
	}
}

func TestRunDeduplicate(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<html><body>content</body></html>`))
	}))
	defer srv.Close()

	s := newTestSpider("")
	// Pass the same URL twice.
	items, err := s.Run([]string{srv.URL, srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 deduplicated item, got %d", len(items))
	}
	if callCount != 1 {
		t.Errorf("expected 1 HTTP call, got %d", callCount)
	}
}
