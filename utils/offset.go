package utils

import (
	"os"
	"strconv"
)

const OffsetFile = ".mailgrid.offset"

// SaveOffset writes the given index to the offset file
func SaveOffset(index int) error {
	return os.WriteFile(OffsetFile, []byte(strconv.Itoa(index)), 0644)
}

// LoadOffset reads the last saved index from the file
func LoadOffset() (int, error) {
	data, err := os.ReadFile(OffsetFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// ResetOffset deletes the offset file
func ResetOffset() error {
	return os.Remove(OffsetFile)
}
