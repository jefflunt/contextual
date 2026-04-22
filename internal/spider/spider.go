package spider

import (
	"container/heap"
	"fmt"
	"regexp"
	"strings"

	"github.com/jluntpcty/contextual/internal/config"
	"github.com/jluntpcty/contextual/internal/fetcher"
	"github.com/jluntpcty/contextual/internal/logger"
	"github.com/jluntpcty/contextual/internal/types"
)

type Direction int

const (
	DirectionSource Direction = iota
	DirectionDown
	DirectionUp
	DirectionSibling
)

type queueEntry struct {
	item           types.Item
	fromConfluence bool
	jumps          int
	direction      Direction
	index          int // required for heap
}

// PriorityQueue implements heap.Interface and holds queueEntries.
type PriorityQueue []*queueEntry

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// 1. Nearness (`jumps` - ascending).
	if pq[i].jumps != pq[j].jumps {
		return pq[i].jumps < pq[j].jumps
	}
	// 2. Direction (Prioritize Down > Up > Sibling).
	if pq[i].direction != pq[j].direction {
		return pq[i].direction < pq[j].direction
	}
	// 3. Discovery Order (tie-breaker).
	return pq[i].index < pq[j].index
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*queueEntry)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

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
	pq := &PriorityQueue{}
	heap.Init(pq)
	visited := map[string]bool{} // key: "<type>:<canonical_url>"

	key := func(item *types.Item) string {
		id := item.ID
		if id == "" {
			id = item.URL
		}
		// Canonicalize URL to prevent circular crawls on same resource.
		canonicalURL := item.URL
		// (Simple canonicalization: remove query params/fragment)
		if idx := strings.Index(canonicalURL, "?"); idx != -1 {
			canonicalURL = canonicalURL[:idx]
		}
		if idx := strings.Index(canonicalURL, "#"); idx != -1 {
			canonicalURL = canonicalURL[:idx]
		}
		return string(item.Type) + ":" + canonicalURL
	}

	enqueue := func(item types.Item, fromConfluence bool, jumps int, dir Direction) {
		k := key(&item)
		if !visited[k] {
			visited[k] = true
			heap.Push(pq, &queueEntry{
				item:           item,
				fromConfluence: fromConfluence,
				jumps:          jumps,
				direction:      dir,
			})
		}
	}

	// Seed from CLI arguments.
	for _, arg := range args {
		item, err := s.ParseItem(arg)
		if err != nil {
			s.logError("%v", err)
			continue
		}
		enqueue(*item, false, 0, DirectionSource)
	}

	host := ""
	email := ""
	token := ""
	maxJumps := 5 // Default
	if s.Config != nil {
		host = s.Config.Atlassian.Host
		email = s.Config.Atlassian.APIUser
		token = s.Config.Atlassian.APIToken
		if s.Config.Atlassian.MaxSpiderJumps > 0 {
			maxJumps = s.Config.Atlassian.MaxSpiderJumps
		}
	}

	var results []types.Item

	for pq.Len() > 0 {
		entry := heap.Pop(pq).(*queueEntry)
		item := entry.item

		if entry.jumps >= maxJumps {
			s.logInfo("Max spider jumps reached for %s, skipping expansion", item.URL)
			results = append(results, item)
			continue
		}

		switch item.Type {
		case types.ItemTypeJira:
			s.logInfo("Fetching jira: %s (jumps: %d)", item.ID, entry.jumps)
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
				s.logInfo("Enqueueing parent: %s", result.ParentKey)
				enqueue(types.Item{
					Type: types.ItemTypeJira,
					ID:   result.ParentKey,
					URL:  fmt.Sprintf("https://%s/browse/%s", host, result.ParentKey),
				}, entry.fromConfluence, entry.jumps+1, DirectionUp)
			}
			for _, st := range result.SubtaskKeys {
				s.logInfo("Enqueueing subtask/linked: %s", st)
				enqueue(types.Item{
					Type: types.ItemTypeJira,
					ID:   st,
					URL:  fmt.Sprintf("https://%s/browse/%s", host, st),
				}, entry.fromConfluence, entry.jumps+1, DirectionDown)
			}
			// Confluence pages and web URLs found in Jira.
			for _, cid := range result.ConfluenceIDs {
				enqueue(types.Item{
					Type: types.ItemTypeConfluence,
					ID:   cid,
					URL:  fmt.Sprintf("https://%s/wiki/rest/api/content/%s", host, cid),
				}, false, entry.jumps+1, DirectionDown)
			}
			for _, u := range result.WebURLs {
				enqueue(types.Item{Type: types.ItemTypeWeb, URL: u}, false, entry.jumps+1, DirectionDown)
			}

		case types.ItemTypeConfluence:
			s.logInfo("Fetching confluence: %s (jumps: %d)", item.ID, entry.jumps)
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
				}, false, entry.jumps+1, DirectionDown)
			}
			// Jira keys found in Confluence (fetch with parent/child expansion).
			for _, jk := range result.JiraKeys {
				enqueue(types.Item{
					Type: types.ItemTypeJira,
					ID:   jk,
					URL:  fmt.Sprintf("https://%s/browse/%s", host, jk),
				}, true, entry.jumps+1, DirectionDown)
			}
			// Web URLs directly linked from Confluence.
			for _, u := range result.WebURLs {
				enqueue(types.Item{Type: types.ItemTypeWeb, URL: u}, false, entry.jumps+1, DirectionDown)
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
