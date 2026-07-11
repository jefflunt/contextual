package writer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jefflunt/contextual/internal/fetcher"
)

type confluenceVersion struct {
	Number int `json:"number"`
}

type confluenceSpace struct {
	Key string `json:"key"`
}

type confluenceStorage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

type confluenceBody struct {
	Storage confluenceStorage `json:"storage"`
}

type confluenceAncestor struct {
	ID string `json:"id"`
}

type confluenceContainer struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Request payloads
type createPageRequest struct {
	Type      string               `json:"type"`
	Title     string               `json:"title"`
	Space     confluenceSpace      `json:"space"`
	Body      confluenceBody       `json:"body"`
	Ancestors []confluenceAncestor `json:"ancestors,omitempty"`
}

type updatePageRequest struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Title   string            `json:"title"`
	Version confluenceVersion `json:"version"`
	Body    confluenceBody    `json:"body"`
}

type createCommentRequest struct {
	Type      string               `json:"type"`
	Container confluenceContainer  `json:"container"`
	Body      confluenceBody       `json:"body"`
	Ancestors []confluenceAncestor `json:"ancestors,omitempty"`
}

// Response structs
type contentResponse struct {
	ID      string            `json:"id"`
	Title   string            `json:"title"`
	Version confluenceVersion `json:"version"`
}

// CreatePage creates a new Confluence page.
func CreatePage(host, email, token, space, title, parentID, bodyHTML string) (string, error) {
	client := fetcher.NewHTTPClient()
	url := fmt.Sprintf("https://%s/wiki/rest/api/content", host)

	reqPayload := createPageRequest{
		Type:  "page",
		Title: title,
		Space: confluenceSpace{Key: space},
		Body: confluenceBody{
			Storage: confluenceStorage{
				Value:          bodyHTML,
				Representation: "storage",
			},
		},
	}

	if parentID != "" {
		reqPayload.Ancestors = []confluenceAncestor{{ID: parentID}}
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("marshalling create page request: %w", err)
	}

	respData, statusCode, err := fetcher.DoRequestWithBody(client, "POST", url, email, token, jsonData)
	if err != nil {
		return "", fmt.Errorf("sending create page request: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", statusCode, string(respData))
	}

	var resp contentResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return "", fmt.Errorf("unmarshalling create page response: %w", err)
	}

	return resp.ID, nil
}

// UpdatePage updates the title and content of an existing Confluence page.
func UpdatePage(host, email, token, pageID, title, bodyHTML string) error {
	client := fetcher.NewHTTPClient()

	// 1. Fetch current version of the page
	fetchURL := fmt.Sprintf("https://%s/wiki/rest/api/content/%s?expand=version", host, pageID)
	respData, statusCode, err := fetcher.DoRequest(client, "GET", fetchURL, email, token)
	if err != nil {
		return fmt.Errorf("fetching current page version: %w", err)
	}
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", statusCode, string(respData))
	}

	var current contentResponse
	if err := json.Unmarshal(respData, &current); err != nil {
		return fmt.Errorf("unmarshalling current version response: %w", err)
	}

	// 2. Put updated content with version incremented
	updateURL := fmt.Sprintf("https://%s/wiki/rest/api/content/%s", host, pageID)
	reqPayload := updatePageRequest{
		ID:    pageID,
		Type:  "page",
		Title: title,
		Version: confluenceVersion{
			Number: current.Version.Number + 1,
		},
		Body: confluenceBody{
			Storage: confluenceStorage{
				Value:          bodyHTML,
				Representation: "storage",
			},
		},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return fmt.Errorf("marshalling update page request: %w", err)
	}

	respData, statusCode, err = fetcher.DoRequestWithBody(client, "PUT", updateURL, email, token, jsonData)
	if err != nil {
		return fmt.Errorf("sending update page request: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", statusCode, string(respData))
	}

	return nil
}

// ReplyToComment posts a reply to an existing comment.
func ReplyToComment(host, email, token, pageID, parentCommentID, bodyText string) (string, error) {
	client := fetcher.NewHTTPClient()
	url := fmt.Sprintf("https://%s/wiki/rest/api/content", host)

	// Ensure the body is wrapped in XHTML paragraph tags if not already done.
	bodyHTML := bodyText
	if !strings.HasPrefix(strings.TrimSpace(bodyHTML), "<") {
		bodyHTML = fmt.Sprintf("<p>%s</p>", bodyText)
	}

	reqPayload := createCommentRequest{
		Type: "comment",
		Container: confluenceContainer{
			ID:   pageID,
			Type: "page",
		},
		Body: confluenceBody{
			Storage: confluenceStorage{
				Value:          bodyHTML,
				Representation: "storage",
			},
		},
		Ancestors: []confluenceAncestor{{ID: parentCommentID}},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("marshalling create comment request: %w", err)
	}

	respData, statusCode, err := fetcher.DoRequestWithBody(client, "POST", url, email, token, jsonData)
	if err != nil {
		return "", fmt.Errorf("sending create comment request: %w", err)
	}

	if statusCode < 200 || statusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", statusCode, string(respData))
	}

	var resp contentResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return "", fmt.Errorf("unmarshalling create comment response: %w", err)
	}

	return resp.ID, nil
}
