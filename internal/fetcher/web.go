package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jefflunt/contextual/internal/types"
)

type WebResult struct {
	Item  types.Item
	Links []string
}

func FetchWeb(rawURL string) (*WebResult, error) {
	client := newHTTPClient()

	body, finalURL, contentType, statusCode, err := fetchFollowingRedirects(client, rawURL, 5)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d for %s", statusCode, rawURL)
	}

	var textContent string
	var links []string

	if strings.Contains(contentType, "text/html") {
		textContent = htmlToText(string(body))
		links = extractHTMLLinks(string(body), "")
	} else {
		textContent = string(body)
	}

	return &WebResult{
		Item: types.Item{
			Type:    types.ItemTypeWeb,
			ID:      "",
			URL:     finalURL,
			Title:   rawURL,
			Content: textContent,
		},
		Links: links,
	}, nil
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func fetchFollowingRedirects(client *http.Client, url string, maxRedirects int) (body []byte, finalURL, contentType string, statusCode int, err error) {
	current := url
	for i := 0; i <= maxRedirects; i++ {
		var req *http.Request
		req, err = http.NewRequest("GET", current, nil)
		if err != nil {
			return
		}

		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil {
			return
		}

		statusCode = resp.StatusCode
		contentType = resp.Header.Get("Content-Type")

		if statusCode >= 300 && statusCode < 400 {
			loc := resp.Header.Get("Location")
			resp.Body.Close()
			if loc == "" {
				err = fmt.Errorf("redirect with no Location header from %s", current)
				return
			}
			current = loc
			continue
		}

		finalURL = current
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		return
	}

	err = fmt.Errorf("too many redirects following %s", url)
	return
}
