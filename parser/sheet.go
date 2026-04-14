package parser

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const sheetTimeout = 30 * time.Second

// GetSheetCSVStream fetches a Google Sheet as a CSV stream.
// The URL must be a Google Sheets URL (docs.google.com) — arbitrary URLs are
// rejected to prevent SSRF. Open redirects are blocked by the HTTP client.
func GetSheetCSVStream(sheetURL string) (io.ReadCloser, error) {
	// Validate the URL is actually a Google Sheets URL before any network I/O.
	parsed, err := url.ParseRequestURI(sheetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid sheet URL: %w", err)
	}
	if parsed.Hostname() != "docs.google.com" {
		return nil, fmt.Errorf("sheet URL must be a Google Sheets URL (docs.google.com), got %q", parsed.Hostname())
	}

	id, gid, err := ExtractSheetInfo(sheetURL)
	if err != nil {
		return nil, err
	}

	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv&gid=%s", id, gid)

	client := &http.Client{
		Timeout: sheetTimeout,
		// Block open redirects — only follow redirects that stay on docs.google.com.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Hostname() != "docs.google.com" {
				return fmt.Errorf("redirect to non-Google domain blocked: %s", req.URL.Hostname())
			}
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Get(exportURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching sheet: %w", err)
	}
	if resp.StatusCode != 200 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("sheet returned status: %s", resp.Status)
	}

	return resp.Body, nil
}
