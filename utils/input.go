package utils

import (
	"os"
	"strings"
)

// ReadTextInput returns the body from inline --text or file path
func ReadTextInput(textArg string) (string, error) {
	if strings.HasSuffix(textArg, ".txt") {
		bytes, err := os.ReadFile(textArg)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	return textArg, nil
}
