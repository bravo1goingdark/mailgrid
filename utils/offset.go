package utils

import (
	"os"
	"strconv"
	"strings"
)

const offsetFile = ".mailgrid.offset"

// LoadOffset returns the saved offset (or 0 if not found or invalid).
func LoadOffset() int {
	data, err := os.ReadFile(offsetFile)
	if err != nil {
		return 0
	}
	offset, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return offset
}

// SaveOffset writes the given offset to the file.
func SaveOffset(offset int) {
	os.WriteFile(offsetFile, []byte(strconv.Itoa(offset)), 0644)
}

// ResetOffset deletes the offset file.
func ResetOffset() {
	os.Remove(offsetFile)
}
