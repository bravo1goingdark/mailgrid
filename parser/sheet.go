package parser

import (
	"fmt"
	"io"
	"mailgrid/utils"
	"net/http"
)

// GetSheetCSVStream fetches a Google Sheet as a CSV stream.
// It extracts the sheet ID and GID from the provided URL and constructs the export URL.
func GetSheetCSVStream(sheetURL string) (io.ReadCloser, error) {
	id, gid, err := utils.ExtractSheetInfo(sheetURL)
	if err != nil {
		return nil, err
	}

	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv&gid=%s", id, gid)
	resp, err := http.Get(exportURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching sheet: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sheet returned status: %s", resp.Status)
	}

	return resp.Body, nil
}
