package fetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/jluntpcty/contextual/internal/types"
)
...
	// If this is an Epic, fetch its child issues.
	log.Printf("Checking if %s is an Epic. Type: %s", issueKey, issue.Fields.Issuetype.Name)
	if issue.Fields.Issuetype.Name == "Epic" {
		searchURL := fmt.Sprintf("https://%s/rest/api/3/search", host)
		jql := fmt.Sprintf(`"Epic Link" = %s OR parent = %s`, issueKey, issueKey)
		
		// Body for search
		jsonBody, _ := json.Marshal(map[string]string{"jql": jql})
		
		data, statusCode, err := doRequestWithBody(client, "POST", searchURL, email, token, jsonBody)
		if err == nil && statusCode >= 200 && statusCode < 300 {
			var searchResp struct {
				Issues []struct {
					Key string `json:"key"`
				} `json:"issues"`
			}
			if json.Unmarshal(data, &searchResp) == nil {
				log.Printf("Found %d issues for Epic %s", len(searchResp.Issues), issueKey)
				for _, iss := range searchResp.Issues {
					subtaskKeys = appendUnique(subtaskKeys, iss.Key)
				}
			} else {
				log.Printf("Failed to unmarshal search response: %v", err)
			}
		} else {
			log.Printf("Search request failed: status %d, err %v", statusCode, err)
		}
	}



func FetchJira(host, email, token, issueKey string) (*JiraResult, error) {
	client := newHTTPClient()

	issueURL := fmt.Sprintf("https://%s/rest/api/3/issue/%s?expand=fields", host, issueKey)
	issueData, statusCode, err := doRequest(client, "GET", issueURL, email, token)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d for %s", statusCode, issueURL)
	}

	var issue jiraIssue
	if err := json.Unmarshal(issueData, &issue); err != nil {
		return nil, fmt.Errorf("parsing jira issue: %w", err)
	}

	// Extract description text and links from ADF.
	descText := extractTextFromADF(issue.Fields.Description)
	allLinks := extractLinksFromADF(issue.Fields.Description, host)

	// Build comments section.
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Status: %s\n", issue.Fields.Status.Name))
	sb.WriteString("\n## Description\n\n")
	sb.WriteString(descText)
	sb.WriteString("\n")

	if len(issue.Fields.Comment.Comments) > 0 {
		sb.WriteString("\n## Comments\n\n")
		for _, c := range issue.Fields.Comment.Comments {
			author := c.Author.DisplayName
			commentText := extractTextFromADF(c.Body)
			sb.WriteString(fmt.Sprintf("**%s**: %s\n\n", author, commentText))
			allLinks = appendUnique(allLinks, extractLinksFromADF(c.Body, host)...)
		}
	}

	// Remote links.
	remoteURL := fmt.Sprintf("https://%s/rest/api/3/issue/%s/remotelink", host, issueKey)
	remoteData, remoteStatus, err := doRequest(client, "GET", remoteURL, email, token)
	if err == nil && remoteStatus >= 200 && remoteStatus < 300 {
		var remoteLinks []jiraRemoteLink
		if json.Unmarshal(remoteData, &remoteLinks) == nil {
			for _, rl := range remoteLinks {
				u := rl.Object.URL
				if u != "" {
					allLinks = appendUnique(allLinks, u)
				}
			}
		}
	}

	var parentKey string
	if issue.Fields.Parent != nil {
		parentKey = issue.Fields.Parent.Key
	}

	subtaskKeys := make([]string, 0, len(issue.Fields.Subtasks)+len(issue.Fields.Issuelinks))
	for _, st := range issue.Fields.Subtasks {
		if key, ok := st["key"].(string); ok {
			subtaskKeys = append(subtaskKeys, key)
		}
	}
	for _, link := range issue.Fields.Issuelinks {
		for _, key := range []string{"outwardIssue", "inwardIssue"} {
			if val, ok := link[key].(map[string]interface{}); ok {
				if issueKey, ok := val["key"].(string); ok {
					subtaskKeys = append(subtaskKeys, issueKey)
				}
			}
		}
	}

	// If this is an Epic, fetch its child issues.
	s.logInfo("Checking if %s is an Epic. Type: %s", issueKey, issue.Fields.Issuetype.Name)
	if issue.Fields.Issuetype.Name == "Epic" {
		searchURL := fmt.Sprintf("https://%s/rest/api/3/search", host)
		jql := fmt.Sprintf(`"Epic Link" = %s OR parent = %s`, issueKey, issueKey)

		// Body for search
		jsonBody, _ := json.Marshal(map[string]string{"jql": jql})

		data, statusCode, err := doRequestWithBody(client, "POST", searchURL, email, token, jsonBody)
		if err == nil && statusCode >= 200 && statusCode < 300 {
			var searchResp struct {
				Issues []struct {
					Key string `json:"key"`
				} `json:"issues"`
			}
			if json.Unmarshal(data, &searchResp) == nil {
				s.logInfo("Found %d issues for Epic %s", len(searchResp.Issues), issueKey)
				for _, iss := range searchResp.Issues {
					subtaskKeys = appendUnique(subtaskKeys, iss.Key)
				}
			} else {
				s.logError("Failed to unmarshal search response: %v", err)
			}
		} else {
			s.logError("Search request failed: status %d, err %v", statusCode, err)
		}
	}

	// Canonical URL
	canonicalURL := fmt.Sprintf("https://%s/browse/%s", host, issueKey)

	// Classify all gathered links.
	var confluenceIDs, jiraKeys, webURLs []string
	for _, u := range allLinks {
		// Resolve relative links.
		if strings.HasPrefix(u, "/") {
			u = "https://" + host + u
		}
		if isConfluenceURL(u, host) {
			if id := extractConfluenceID(u); id != "" {
				confluenceIDs = appendUnique(confluenceIDs, id)
			}
		} else if strings.Contains(u, "/browse/") {
			if key := extractJiraKeyFromURL(u); key != "" {
				jiraKeys = appendUnique(jiraKeys, key)
			}
		} else {
			webURLs = appendUnique(webURLs, u)
		}
	}

	return &JiraResult{
		Item: types.Item{
			Type:    types.ItemTypeJira,
			ID:      issueKey,
			URL:     canonicalURL,
			Title:   issue.Fields.Summary,
			Content: sb.String(),
		},
		ParentKey:     parentKey,
		SubtaskKeys:   append(subtaskKeys, jiraKeys...),
		ConfluenceIDs: confluenceIDs,
		WebURLs:       webURLs,
	}, nil
}

func extractJiraKeyFromURL(u string) string {
	parts := strings.Split(u, "/browse/")
	if len(parts) < 2 {
		return ""
	}
	seg := parts[len(parts)-1]
	// Strip query/fragment.
	for _, sep := range []string{"?", "#"} {
		if idx := strings.Index(seg, sep); idx != -1 {
			seg = seg[:idx]
		}
	}
	seg = strings.TrimRight(seg, "/")
	if regexp.MustCompile(`^[A-Z]+-\d+$`).MatchString(seg) {
		return seg
	}
	return ""
}

// extractTextFromADF recursively walks an Atlassian Document Format (ADF) node
// and returns Markdown-formatted text.
func extractTextFromADF(node interface{}) string {
	return adfToMarkdown(node, 0, false, 0)
}

// adfToMarkdown converts an ADF node tree to Markdown.
// depth is the current list nesting depth (0 = top level).
// inOrderedList indicates the immediate parent list is ordered.
// listIndex is the 1-based index of this item within an ordered list.
func adfToMarkdown(node interface{}, depth int, inOrderedList bool, listIndex int) string {
	if node == nil {
		return ""
	}
	switch v := node.(type) {
	case string:
		return v
	case map[string]interface{}:
		nodeType, _ := v["type"].(string)

		switch nodeType {
		case "text":
			text, _ := v["text"].(string)
			if text == "" {
				return ""
			}
			// Apply marks: bold, italic, code, link, strikethrough.
			marks, _ := v["marks"].([]interface{})
			for _, m := range marks {
				mMap, ok := m.(map[string]interface{})
				if !ok {
					continue
				}
				markType, _ := mMap["type"].(string)
				switch markType {
				case "strong":
					text = "**" + text + "**"
				case "em":
					text = "_" + text + "_"
				case "code":
					text = "`" + text + "`"
				case "strike":
					text = "~~" + text + "~~"
				case "link":
					if attrs, ok := mMap["attrs"].(map[string]interface{}); ok {
						if href, _ := attrs["href"].(string); href != "" {
							text = "[" + text + "](" + href + ")"
						}
					}
				}
			}
			return text

		case "hardBreak":
			return "\n"

		case "paragraph":
			return adfChildrenInline(v) + "\n\n"

		case "heading":
			attrs, _ := v["attrs"].(map[string]interface{})
			level := 1
			if l, ok := attrs["level"].(float64); ok {
				level = int(l)
			}
			prefix := strings.Repeat("#", level)
			return prefix + " " + adfChildrenInline(v) + "\n\n"

		case "bulletList":
			return adfList(v, depth, false) + "\n"

		case "orderedList":
			return adfList(v, depth, true) + "\n"

		case "listItem":
			indent := strings.Repeat("  ", depth)
			var prefix string
			if inOrderedList {
				prefix = fmt.Sprintf("%s%d. ", indent, listIndex)
			} else {
				prefix = indent + "- "
			}
			// Collect child content; nested lists are rendered inline.
			var parts []string
			if content, ok := v["content"].([]interface{}); ok {
				for _, child := range content {
					childMap, ok := child.(map[string]interface{})
					if !ok {
						continue
					}
					childType, _ := childMap["type"].(string)
					switch childType {
					case "bulletList":
						parts = append(parts, "\n"+adfList(child, depth+1, false))
					case "orderedList":
						parts = append(parts, "\n"+adfList(child, depth+1, true))
					default:
						if t := strings.TrimSpace(adfToMarkdown(child, depth, false, 0)); t != "" {
							parts = append(parts, t)
						}
					}
				}
			}
			return prefix + strings.Join(parts, " ") + "\n"

		case "codeBlock":
			attrs, _ := v["attrs"].(map[string]interface{})
			lang, _ := attrs["language"].(string)
			var code strings.Builder
			if content, ok := v["content"].([]interface{}); ok {
				for _, child := range content {
					code.WriteString(adfToMarkdown(child, depth, false, 0))
				}
			}
			return "```" + lang + "\n" + code.String() + "\n```\n\n"

		case "blockquote":
			inner := adfChildrenBlock(v, depth)
			var out strings.Builder
			for _, line := range strings.Split(strings.TrimRight(inner, "\n"), "\n") {
				out.WriteString("> " + line + "\n")
			}
			return out.String() + "\n"

		case "rule":
			return "---\n\n"

		case "media", "mediaSingle", "mediaGroup":
			// Emit alt text or filename if available.
			if attrs, ok := v["attrs"].(map[string]interface{}); ok {
				if alt, _ := attrs["alt"].(string); alt != "" {
					return "_[image: " + alt + "]_\n\n"
				}
			}
			return ""

		case "mention":
			if attrs, ok := v["attrs"].(map[string]interface{}); ok {
				if text, _ := attrs["text"].(string); text != "" {
					return "@" + strings.TrimPrefix(text, "@")
				}
			}
			return ""

		case "emoji":
			if attrs, ok := v["attrs"].(map[string]interface{}); ok {
				if short, _ := attrs["shortName"].(string); short != "" {
					return short
				}
				if text, _ := attrs["text"].(string); text != "" {
					return text
				}
			}
			return ""

		case "date":
			if attrs, ok := v["attrs"].(map[string]interface{}); ok {
				if ts, _ := attrs["timestamp"].(string); ts != "" {
					return ts
				}
			}
			return ""

		case "status":
			if attrs, ok := v["attrs"].(map[string]interface{}); ok {
				if text, _ := attrs["text"].(string); text != "" {
					return "`" + text + "`"
				}
			}
			return ""

		case "table":
			return adfTable(v, depth)

		case "expand", "nestedExpand":
			var title string
			if attrs, ok := v["attrs"].(map[string]interface{}); ok {
				title, _ = attrs["title"].(string)
			}
			inner := adfChildrenBlock(v, depth)
			if title != "" {
				return "**" + title + "**\n\n" + inner
			}
			return inner

		case "panel":
			inner := adfChildrenBlock(v, depth)
			var out strings.Builder
			for _, line := range strings.Split(strings.TrimRight(inner, "\n"), "\n") {
				out.WriteString("> " + line + "\n")
			}
			return out.String() + "\n"

		default:
			// For unknown block nodes, recurse into children with block separation.
			return adfChildrenBlock(v, depth)
		}

	case []interface{}:
		var parts []string
		for _, item := range v {
			if t := adfToMarkdown(item, depth, false, 0); t != "" {
				parts = append(parts, t)
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// adfChildrenInline collects all children as inline text (no block separators).
func adfChildrenInline(v map[string]interface{}) string {
	var parts []string
	if content, ok := v["content"].([]interface{}); ok {
		for _, child := range content {
			if t := adfToMarkdown(child, 0, false, 0); t != "" {
				parts = append(parts, t)
			}
		}
	}
	return strings.Join(parts, "")
}

// adfChildrenBlock collects block children separated by newlines.
func adfChildrenBlock(v map[string]interface{}, depth int) string {
	content, ok := v["content"].([]interface{})
	if !ok {
		return ""
	}
	var out strings.Builder
	for _, child := range content {
		out.WriteString(adfToMarkdown(child, depth, false, 0))
	}
	return out.String()
}

// adfList renders a bulletList or orderedList node.
func adfList(node interface{}, depth int, ordered bool) string {
	v, ok := node.(map[string]interface{})
	if !ok {
		return ""
	}
	content, ok := v["content"].([]interface{})
	if !ok {
		return ""
	}
	var out strings.Builder
	for i, child := range content {
		out.WriteString(adfToMarkdown(child, depth, ordered, i+1))
	}
	return out.String()
}

// adfTable renders a table node as a Markdown table.
func adfTable(v map[string]interface{}, depth int) string {
	content, ok := v["content"].([]interface{})
	if !ok {
		return ""
	}
	var rows [][]string
	for _, rowNode := range content {
		rowMap, ok := rowNode.(map[string]interface{})
		if !ok {
			continue
		}
		cells, ok := rowMap["content"].([]interface{})
		if !ok {
			continue
		}
		var row []string
		for _, cellNode := range cells {
			cellMap, ok := cellNode.(map[string]interface{})
			if !ok {
				row = append(row, "")
				continue
			}
			row = append(row, strings.TrimSpace(adfChildrenBlock(cellMap, depth)))
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return ""
	}
	// Normalise column count.
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
	var out strings.Builder
	for i, row := range rows {
		for j, cell := range row {
			cell = strings.ReplaceAll(cell, "\n", " ")
			row[j] = cell
		}
		out.WriteString("| " + strings.Join(row, " | ") + " |\n")
		if i == 0 {
			sep := make([]string, cols)
			for j := range sep {
				sep[j] = "---"
			}
			out.WriteString("| " + strings.Join(sep, " | ") + " |\n")
		}
	}
	return out.String() + "\n"
}

// extractLinksFromADF walks ADF and collects all links.
func extractLinksFromADF(node interface{}, host string) []string {
	var links []string
	if node == nil {
		return links
	}
	switch v := node.(type) {
	case map[string]interface{}:
		// Check for links in marks.
		if marks, ok := v["marks"].([]interface{}); ok {
			for _, m := range marks {
				mMap, ok := m.(map[string]interface{})
				if !ok {
					continue
				}
				if mMap["type"] == "link" {
					if attrs, ok := mMap["attrs"].(map[string]interface{}); ok {
						href, _ := attrs["href"].(string)
						if href != "" {
							links = appendUnique(links, href)
						}
					}
				}
			}
		}
		// Check attrs for href/url fields.
		if attrs, ok := v["attrs"].(map[string]interface{}); ok {
			for _, key := range []string{"href", "url"} {
				href, _ := attrs[key].(string)
				if href != "" {
					links = appendUnique(links, href)
				}
			}
		}
		// Recurse into content.
		if content, ok := v["content"].([]interface{}); ok {
			for _, child := range content {
				links = appendUnique(links, extractLinksFromADF(child, host)...)
			}
		}
	case []interface{}:
		for _, item := range v {
			links = appendUnique(links, extractLinksFromADF(item, host)...)
		}
	}
	return links
}

var confluencePagePathRe = regexp.MustCompile(`/wiki/spaces/[^/]+/pages/(\d+)`)
var confluenceContentRe = regexp.MustCompile(`/wiki/rest/api/content/(\d+)`)

func extractConfluenceID(u string) string {
	if m := confluencePagePathRe.FindStringSubmatch(u); len(m) > 1 {
		return m[1]
	}
	if m := confluenceContentRe.FindStringSubmatch(u); len(m) > 1 {
		return m[1]
	}
	return ""
}

func isConfluenceURL(u, host string) bool {
	return strings.Contains(u, host+"/wiki/")
}

func isAtlassianURL(u, host string) bool {
	return strings.Contains(u, host)
}

func appendUnique(dst []string, items ...string) []string {
	seen := make(map[string]bool, len(dst))
	for _, s := range dst {
		seen[s] = true
	}
	for _, s := range items {
		if s != "" && !seen[s] {
			seen[s] = true
			dst = append(dst, s)
		}
	}
	return dst
}

func doRequest(client *http.Client, method, url, email, token string) ([]byte, int, error) {
	return doRequestWithBody(client, method, url, email, token, nil)
}

func doRequestWithBody(client *http.Client, method, url, email, token string, body []byte) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if email != "" && token != "" {
		req.SetBasicAuth(email, token)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

// Jira API response types.

type jiraIssue struct {
	Key    string     `json:"key"`
	Fields jiraFields `json:"fields"`
}

type jiraFields struct {
	Summary     string                   `json:"summary"`
	Description interface{}              `json:"description"`
	Status      jiraStatus               `json:"status"`
	Parent      *jiraParent              `json:"parent"`
	Subtasks    []map[string]interface{} `json:"subtasks"`
	Issuelinks  []map[string]interface{} `json:"issuelinks"`
	Comment     jiraComments             `json:"comment"`
	Issuetype   jiraIssuetype            `json:"issuetype"`
}

type jiraIssuetype struct {
	Name string `json:"name"`
}

type jiraStatus struct {
	Name string `json:"name"`
}

type jiraParent struct {
	Key string `json:"key"`
}

type jiraComments struct {
	Comments []jiraComment `json:"comments"`
}

type jiraComment struct {
	Author jiraUser    `json:"author"`
	Body   interface{} `json:"body"`
}

type jiraUser struct {
	DisplayName string `json:"displayName"`
}

type jiraRemoteLink struct {
	Object jiraRemoteLinkObject `json:"object"`
}

type jiraRemoteLinkObject struct {
	URL string `json:"url"`
}
