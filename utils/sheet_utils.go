package utils

import (
	"errors"
	"regexp"
)

// precompiled regular expression for matching Google Sheets URLs and extracting document ID.
var sheetURLRegex = regexp.MustCompile(`spreadsheets/d/([a-zA-Z0-9-_]+)`)

// precompiled regular expressions for extracting GID from different URL formats
var gidFragmentRegex = regexp.MustCompile(`#gid=([^&]*)`)
var gidQueryRegex = regexp.MustCompile(`[?&]gid=([^&]*)`)

// ExtractSheetInfo extracts the Google Sheets document ID and GID (sheet/tab ID)
// from a given URL.
//
// Supported Google Sheets URL formats:
//   - https://docs.google.com/spreadsheets/d/1a2b3c4d5e6f7g8h9i0j/edit#gid=123456789
//   - https://docs.google.com/spreadsheets/d/1a2b3c4d5e6f7g8h9i0j/export?format=csv&gid=987654321
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
	// Extract document ID
	matches := sheetURLRegex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", "", errors.New("invalid Google Sheets URL format")
	}
	id := matches[1]

	// Extract GID from fragment (e.g., #gid=123456789)
	gid := "0"
	gidMatches := gidFragmentRegex.FindStringSubmatch(url)
	if len(gidMatches) >= 2 && gidMatches[1] != "" {
		gid = gidMatches[1]
	} else {
		// Extract GID from query parameters (e.g., ?gid=987654321 or &gid=987654321)
		gidMatches = gidQueryRegex.FindStringSubmatch(url)
		if len(gidMatches) >= 2 && gidMatches[1] != "" {
			gid = gidMatches[1]
		}
	}

	return id, gid, nil
}
