package parser

import (
	"errors"
	"regexp"
)

func ExtractSheetInfo(url string) (string, string, error) {
	re := regexp.MustCompile(`spreadsheets/d/([a-zA-Z0-9-_]+)(?:/[^#]*)?(?:#gid=(\d+))?`)
	matches := re.FindStringSubmatch(url)

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
