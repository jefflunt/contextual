package spider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jluntpcty/contextual/internal/config"
	"github.com/jluntpcty/contextual/internal/fetcher"
	"github.com/jluntpcty/contextual/internal/logger"
	"github.com/jluntpcty/contextual/internal/types"
)

var (
	jiraKeyRe      = regexp.MustCompile(`^[A-Z]+-\d+$`)
	confluenceIDRe = regexp.MustCompile(`^\d{8,}$`)
	numericIDRe    = regexp.MustCompile(`/(\d{8,})(?:[/?#]|$)`)
)

type Spider struct {
	Config *config.Config
	Log    *logger.Logger
}

func New(cfg *config.Config, log *logger.Logger) *Spider {
	return &Spider{Config: cfg, Log: log}
}

func (s *Spider) logInfo(format string, args ...interface{}) {
	if s.Log != nil {
		s.Log.Info(format, args...)
	}
}

func (s *Spider) logError(format string, args ...interface{}) {
	if s.Log != nil {
		s.Log.Error(format, args...)
	}
}

// ParseItem classifies a CLI argument into an Item.
func (s *Spider) ParseItem(arg string) (*types.Item, error) {
	host := ""
	if s.Config != nil {
		host = s.Config.Atlassian.Host
	}

	// 1. Jira issue key: e.g. CTX-1234
	if jiraKeyRe.MatchString(arg) {
		u := ""
		if host != "" {
			u = fmt.Sprintf("https://%s/browse/%s", host, arg)
		}
		return &types.Item{Type: types.ItemTypeJira, ID: arg, URL: u}, nil
	}

	// 2. Pure numeric 8+ digit → Confluence page ID
	if confluenceIDRe.MatchString(arg) {
		u := ""
		if host != "" {
			u = fmt.Sprintf("https://%s/wiki/rest/api/content/%s", host, arg)
		}
		return &types.Item{Type: types.ItemTypeConfluence, ID: arg, URL: u}, nil
	}

	if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
		// 3. Atlassian Jira URL: contains <host>/browse/
		if host != "" && strings.Contains(arg, host+"/browse/") {
			key := extractJiraKeyFromURL(arg)
			if key != "" {
				return &types.Item{Type: types.ItemTypeJira, ID: key, URL: arg}, nil
			}
		}

		// 4. Atlassian Confluence URL: contains <host>/wiki/
		if host != "" && strings.Contains(arg, host+"/wiki/") {
			id := extractConfluenceIDFromURL(arg)
			if id != "" {
				return &types.Item{Type: types.ItemTypeConfluence, ID: id, URL: arg}, nil
			}
		}

		// 5. Generic web URL
		return &types.Item{Type: types.ItemTypeWeb, ID: "", URL: arg}, nil
	}

	return nil, fmt.Errorf("unrecognised argument: %q", arg)
}

func (s *Spider) Run(args []string) ([]types.Item, error) {
	type queueEntry struct {
		item           types.Item
		fromConfluence bool // true when this Jira item was discovered via Confluence
	}

	var queue []queueEntry
	visited := map[string]bool{} // key: "<type>:<id_or_url>"

	key := func(item *types.Item) string {
		id := item.ID
		if id == "" {
			id = item.URL
		}
		return string(item.Type) + ":" + id
	}

	enqueue := func(item types.Item, fromConfluence bool) {
		k := key(&item)
		if !visited[k] {
			visited[k] = true
			queue = append(queue, queueEntry{item, fromConfluence})
		}
	}

	// Seed from CLI arguments.
	for _, arg := range args {
		item, err := s.ParseItem(arg)
		if err != nil {
			s.logError("%v", err)
			continue
		}
		enqueue(*item, false)
	}

	host := ""
	email := ""
	token := ""
	if s.Config != nil {
		host = s.Config.Atlassian.Host
		email = s.Config.Atlassian.APIUser
		token = s.Config.Atlassian.APIToken
	}

	var results []types.Item

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		item := entry.item

		switch item.Type {
		case types.ItemTypeJira:
			s.logInfo("Fetching jira: %s", item.ID)
			if host == "" {
				s.logError("atlassian.host not configured, cannot fetch Jira item %s — see contextual.config.example.yml", item.ID)
				continue
			}
			result, err := fetcher.FetchJira(host, email, token, item.ID)
			if err != nil {
				s.logError("Failed to fetch jira %s: %v", item.ID, err)
				continue
			}
			results = append(results, result.Item)

			// Always fetch parent and subtasks.
			if result.ParentKey != "" {
				enqueue(types.Item{
					Type: types.ItemTypeJira,
					ID:   result.ParentKey,
					URL:  fmt.Sprintf("https://%s/browse/%s", host, result.ParentKey),
				}, entry.fromConfluence)
			}
			for _, sk := range result.SubtaskKeys {
				enqueue(types.Item{
					Type: types.ItemTypeJira,
					ID:   sk,
					URL:  fmt.Sprintf("https://%s/browse/%s", host, sk),
				}, entry.fromConfluence)
			}
			// Confluence pages and web URLs found in Jira.
			for _, cid := range result.ConfluenceIDs {
				enqueue(types.Item{
					Type: types.ItemTypeConfluence,
					ID:   cid,
					URL:  fmt.Sprintf("https://%s/wiki/rest/api/content/%s", host, cid),
				}, false)
			}
			for _, u := range result.WebURLs {
				enqueue(types.Item{Type: types.ItemTypeWeb, URL: u}, false)
			}

		case types.ItemTypeConfluence:
			s.logInfo("Fetching confluence: %s", item.ID)
			if host == "" {
				s.logError("atlassian.host not configured, cannot fetch Confluence item %s — see contextual.config.example.yml", item.ID)
				continue
			}
			result, err := fetcher.FetchConfluence(host, email, token, item.ID)
			if err != nil {
				s.logError("Failed to fetch confluence %s: %v", item.ID, err)
				continue
			}
			results = append(results, result.Item)

			// Child pages.
			for _, cid := range result.ChildIDs {
				enqueue(types.Item{
					Type: types.ItemTypeConfluence,
					ID:   cid,
					URL:  fmt.Sprintf("https://%s/wiki/rest/api/content/%s", host, cid),
				}, false)
			}
			// Jira keys found in Confluence (fetch with parent/child expansion).
			for _, jk := range result.JiraKeys {
				enqueue(types.Item{
					Type: types.ItemTypeJira,
					ID:   jk,
					URL:  fmt.Sprintf("https://%s/browse/%s", host, jk),
				}, true)
			}
			// Web URLs directly linked from Confluence.
			for _, u := range result.WebURLs {
				enqueue(types.Item{Type: types.ItemTypeWeb, URL: u}, false)
			}

		case types.ItemTypeWeb:
			s.logInfo("Fetching web: %s", item.URL)
			result, err := fetcher.FetchWeb(item.URL)
			if err != nil {
				s.logError("Failed to fetch %s: %v", item.URL, err)
				continue
			}
			results = append(results, result.Item)
			// Do NOT spider further links found on web pages.
		}
	}

	return results, nil
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
	if jiraKeyRe.MatchString(seg) {
		return seg
	}
	return ""
}

func extractConfluenceIDFromURL(u string) string {
	if m := numericIDRe.FindStringSubmatch(u); len(m) > 1 {
		return m[1]
	}
	return ""
}
