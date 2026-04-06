package types

type ItemType string

const (
	ItemTypeJira       ItemType = "jira"
	ItemTypeConfluence ItemType = "confluence"
	ItemTypeWeb        ItemType = "web"
)

type Item struct {
	Type    ItemType
	ID      string // Jira issue key or Confluence page ID
	URL     string // canonical URL
	Title   string
	Content string // formatted content
}
