package parser

import "strings"

// ParseFilterString converts a comma-separated key=value string into a map.
func ParseFilterString(f string) map[string]string {
	result := make(map[string]string, 8)

	f = strings.TrimSpace(f)
	if f == "" {
		return result
	}

	for _, pair := range strings.Split(f, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		if eq := strings.IndexByte(pair, '='); eq != -1 {
			key := strings.ToLower(strings.TrimSpace(pair[:eq]))
			value := strings.TrimSpace(pair[eq+1:])
			if key != "" {
				result[key] = value
			}
		}
	}
	return result
}
