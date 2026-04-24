package fetcher

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jefflunt/contextual/internal/types"
	"golang.org/x/net/html"
)

type ConfluenceResult struct {
	Item     types.Item
	ChildIDs []string
	JiraKeys []string
	WebURLs  []string
}

var jiraKeyRe = regexp.MustCompile(`\b[A-Z]+-\d+\b`)

func FetchConfluence(host, email, token, pageID string) (*ConfluenceResult, error) {
	client := newHTTPClient()

	pageURL := fmt.Sprintf("https://%s/wiki/rest/api/content/%s?expand=body.storage,children.page,space,ancestors", host, pageID)
	data, statusCode, err := doRequest(client, "GET", pageURL, email, token)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d for %s", statusCode, pageURL)
	}

	var page confluencePage
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("parsing confluence page: %w", err)
	}

	storageXHTML := page.Body.Storage.Value
	textContent := htmlToText(storageXHTML)
	allLinks := extractHTMLLinks(storageXHTML, host)
	jiraKeys := extractJiraKeys(storageXHTML)

	// Build webURLs with all links.
	var webURLs []string
	for _, u := range allLinks {
		// Resolve relative links.
		if strings.HasPrefix(u, "/") {
			u = "https://" + host + u
		}
		webURLs = appendUnique(webURLs, u)
	}

	var sb strings.Builder
	sb.WriteString(textContent)

	var childIDs []string
	for _, child := range page.Children.Page.Results {
		childIDs = append(childIDs, child.ID)
	}

	canonicalURL := fmt.Sprintf("https://%s/wiki/rest/api/content/%s", host, pageID)
	return &ConfluenceResult{
		Item: types.Item{
			Type:    types.ItemTypeConfluence,
			ID:      pageID,
			URL:     canonicalURL,
			Title:   page.Title,
			Content: sb.String(),
		},
		ChildIDs: childIDs,
		JiraKeys: jiraKeys,
		WebURLs:  webURLs,
	}, nil
}

func htmlToText(rawHTML string) string {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return rawHTML
	}
	var buf strings.Builder
	htmlNodeToMarkdown(doc, &buf, 0, false, 0)
	return strings.TrimSpace(buf.String())
}

// htmlNodeToMarkdown walks an HTML node tree and writes Markdown to buf.
// depth is the list nesting level, ordered indicates an ordered list parent,
// and listIndex is the 1-based position within an ordered list.
func htmlNodeToMarkdown(n *html.Node, buf *strings.Builder, depth int, ordered bool, listIndex int) {
	if n.Type == html.ElementNode {
		tag := strings.ToLower(n.Data)
		switch tag {
		case "script", "style":
			return // skip entirely

		case "h1", "h2", "h3", "h4", "h5", "h6":
			level := int(tag[1] - '0')
			buf.WriteString(strings.Repeat("#", level) + " ")
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				htmlInlineToMarkdown(c, buf)
			}
			buf.WriteString("\n\n")
			return

		case "p":
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				htmlInlineToMarkdown(c, buf)
			}
			buf.WriteString("\n\n")
			return

		case "br":
			buf.WriteString("\n")
			return

		case "hr":
			buf.WriteString("---\n\n")
			return

		case "ul":
			for i, c := 0, n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && strings.ToLower(c.Data) == "li" {
					i++
					htmlNodeToMarkdown(c, buf, depth, false, i)
				}
			}
			if depth == 0 {
				buf.WriteString("\n")
			}
			return

		case "ol":
			for i, c := 0, n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && strings.ToLower(c.Data) == "li" {
					i++
					htmlNodeToMarkdown(c, buf, depth, true, i)
				}
			}
			if depth == 0 {
				buf.WriteString("\n")
			}
			return

		case "li":
			indent := strings.Repeat("  ", depth)
			var prefix string
			if ordered {
				prefix = fmt.Sprintf("%s%d. ", indent, listIndex)
			} else {
				prefix = indent + "- "
			}
			buf.WriteString(prefix)
			// Inline content first, then nested lists.
			var nestedBuf strings.Builder
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode {
					childTag := strings.ToLower(c.Data)
					if childTag == "ul" || childTag == "ol" {
						nestedBuf.WriteString("\n")
						for i, gc := 0, c.FirstChild; gc != nil; gc = gc.NextSibling {
							if gc.Type == html.ElementNode && strings.ToLower(gc.Data) == "li" {
								i++
								htmlNodeToMarkdown(gc, &nestedBuf, depth+1, childTag == "ol", i)
							}
						}
						continue
					}
				}
				htmlInlineToMarkdown(c, buf)
			}
			buf.WriteString(nestedBuf.String())
			buf.WriteString("\n")
			return

		case "pre", "code":
			// Fenced code block for <pre>, inline code for standalone <code>.
			if tag == "pre" {
				// Check for inner <code> with a language class.
				lang := ""
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && strings.ToLower(c.Data) == "code" {
						for _, attr := range c.Attr {
							if attr.Key == "class" {
								for _, cls := range strings.Fields(attr.Val) {
									if strings.HasPrefix(cls, "language-") {
										lang = strings.TrimPrefix(cls, "language-")
									}
								}
							}
						}
					}
				}
				buf.WriteString("```" + lang + "\n")
				var codeBuf strings.Builder
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					collectText(c, &codeBuf)
				}
				buf.WriteString(codeBuf.String())
				buf.WriteString("\n```\n\n")
				return
			}
			// Inline <code> — handled in htmlInlineToMarkdown.

		case "blockquote":
			var inner strings.Builder
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				htmlNodeToMarkdown(c, &inner, depth, false, 0)
			}
			for _, line := range strings.Split(strings.TrimRight(inner.String(), "\n"), "\n") {
				buf.WriteString("> " + line + "\n")
			}
			buf.WriteString("\n")
			return

		case "table":
			htmlTableToMarkdown(n, buf)
			return

		case "thead", "tbody", "tfoot", "tr", "td", "th":
			// Handled by htmlTableToMarkdown; skip if encountered top-level.
			return
		}
	}

	if n.Type == html.TextNode {
		text := n.Data
		// Normalise whitespace for non-pre contexts.
		if trimmed := strings.TrimSpace(text); trimmed != "" {
			buf.WriteString(trimmed + "\n")
		}
		return
	}

	// Default: recurse into children.
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		htmlNodeToMarkdown(c, buf, depth, ordered, listIndex)
	}
}

// htmlInlineToMarkdown handles inline elements (bold, italic, code, links, text).
func htmlInlineToMarkdown(n *html.Node, buf *strings.Builder) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
		return
	}
	if n.Type != html.ElementNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlInlineToMarkdown(c, buf)
		}
		return
	}
	tag := strings.ToLower(n.Data)
	switch tag {
	case "strong", "b":
		buf.WriteString("**")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlInlineToMarkdown(c, buf)
		}
		buf.WriteString("**")
	case "em", "i":
		buf.WriteString("_")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlInlineToMarkdown(c, buf)
		}
		buf.WriteString("_")
	case "code":
		buf.WriteString("`")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collectText(c, buf)
		}
		buf.WriteString("`")
	case "a":
		var href string
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				href = attr.Val
				break
			}
		}
		var text strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlInlineToMarkdown(c, &text)
		}
		linkText := text.String()
		if href != "" && linkText != "" {
			buf.WriteString("[" + linkText + "](" + href + ")")
		} else if linkText != "" {
			buf.WriteString(linkText)
		}
	case "br":
		buf.WriteString("\n")
	case "script", "style":
		return
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			htmlInlineToMarkdown(c, buf)
		}
	}
}

// collectText extracts raw text from a node tree without any Markdown formatting.
func collectText(n *html.Node, buf *strings.Builder) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectText(c, buf)
	}
}

// htmlTableToMarkdown converts an HTML <table> node to a Markdown table.
func htmlTableToMarkdown(n *html.Node, buf *strings.Builder) {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if tag == "tr" {
				var row []string
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode {
						ct := strings.ToLower(c.Data)
						if ct == "td" || ct == "th" {
							var cellBuf strings.Builder
							for gc := c.FirstChild; gc != nil; gc = gc.NextSibling {
								htmlInlineToMarkdown(gc, &cellBuf)
							}
							row = append(row, strings.TrimSpace(cellBuf.String()))
						}
					}
				}
				if len(row) > 0 {
					rows = append(rows, row)
				}
				return
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)

	if len(rows) == 0 {
		return
	}
	cols := 0
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	for i := range rows {
		for len(rows[i]) < cols {
			rows[i] = append(rows[i], "")
		}
	}
	for i, row := range rows {
		buf.WriteString("| " + strings.Join(row, " | ") + " |\n")
		if i == 0 {
			sep := make([]string, cols)
			for j := range sep {
				sep[j] = "---"
			}
			buf.WriteString("| " + strings.Join(sep, " | ") + " |\n")
		}
	}
	buf.WriteString("\n")
}

func extractHTMLLinks(rawHTML, host string) []string {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil
	}
	var links []string
	seen := map[string]bool{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.ToLower(n.Data) == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					if href == "" || seen[href] {
						break
					}
					seen[href] = true
					links = append(links, href)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return links
}

func extractJiraKeys(content string) []string {
	matches := jiraKeyRe.FindAllString(content, -1)
	seen := map[string]bool{}
	var keys []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			keys = append(keys, m)
		}
	}
	return keys
}

// Confluence API response types.

type confluencePage struct {
	ID       string             `json:"id"`
	Title    string             `json:"title"`
	Body     confluenceBody     `json:"body"`
	Children confluenceChildren `json:"children"`
}

type confluenceBody struct {
	Storage confluenceStorage `json:"storage"`
}

type confluenceStorage struct {
	Value string `json:"value"`
}

type confluenceChildren struct {
	Page confluencePageList `json:"page"`
}

type confluencePageList struct {
	Results []confluencePageRef `json:"results"`
}

type confluencePageRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
