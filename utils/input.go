package utils

import (
	"os"
)

// ReadTextInput returns the body from inline --text or file path.
// It checks if the argument is an existing file path, otherwise treats it as inline text.
func ReadTextInput(textArg string) (string, error) {
	if info, err := os.Stat(textArg); err == nil && !info.IsDir() {
		bytes, err := os.ReadFile(textArg)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
	return textArg, nil
}
