package parser

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const sheetTimeout = 30 * time.Second

// GetSheetCSVStream fetches a Google Sheet as a CSV stream.
// It extracts of sheet ID and GID from provided URL and constructs export URL.
func GetSheetCSVStream(sheetURL string) (io.ReadCloser, error) {
	id, gid, err := ExtractSheetInfo(sheetURL)
	if err != nil {
		return nil, err
	}

	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv&gid=%s", id, gid)

	client := &http.Client{
		Timeout: sheetTimeout,
	}

	resp, err := client.Get(exportURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching sheet: %w", err)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("sheet returned status: %s", resp.Status)
	}

	return resp.Body, nil
}
