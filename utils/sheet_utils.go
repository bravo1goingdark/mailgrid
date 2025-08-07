package utils

import (
	"errors"
	"regexp"
)

// precompiled regular expression for matching Google Sheets URLs.
var sheetURLRegex = regexp.MustCompile(`spreadsheets/d/([a-zA-Z0-9-_]+)(?:/[^#]*)?(?:#gid=(\d+))?`)

// ExtractSheetInfo extracts the Google Sheets document ID and GID (sheet/tab ID)
// from a given URL.
//
// Supported Google Sheets URL formats:
//   - https://docs.google.com/spreadsheets/d/1a2b3c4d5e6f7g8h9i0j/edit#gid=123456789
//   - https://docs.google.com/spreadsheets/d/1a2b3c4d5e6f7g8h9i0j
//
// Parameters:
//   - url: The full URL of a Google Sheets document.
//
// Returns:
//   - id: The document ID (string)
//   - gid: The sheet/tab GID (string). Defaults to "0" if not found
//   - error: If the URL format is invalid
func ExtractSheetInfo(url string) (string, string, error) {
	matches := sheetURLRegex.FindStringSubmatch(url)

	if len(matches) < 2 {
		return "", "", errors.New("invalid Google Sheets URL format")
	}
	id := matches[1]
	gid := "0"
	if len(matches) >= 3 && matches[2] != "" {
		gid = matches[2]
	}
	return id, gid, nil
}
