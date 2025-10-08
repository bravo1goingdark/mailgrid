package utils_test

import (
	"testing"

	"github.com/bravo1goingdark/mailgrid/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractSheetInfo_ValidURLs(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedID  string
		expectedGID string
		expectedErr bool
	}{
		{
			name:        "valid sheet URL with GID",
			url:         "https://docs.google.com/spreadsheets/d/1ABC123DEF456/edit#gid=123456789",
			expectedID:  "1ABC123DEF456",
			expectedGID: "123456789",
			expectedErr: false,
		},
		{
			name:        "valid sheet URL without GID",
			url:         "https://docs.google.com/spreadsheets/d/1ABC123DEF456/edit",
			expectedID:  "1ABC123DEF456",
			expectedGID: "0",
			expectedErr: false,
		},
		{
			name:        "sheet URL with export format",
			url:         "https://docs.google.com/spreadsheets/d/1ABC123DEF456/export?format=csv&gid=987654321",
			expectedID:  "1ABC123DEF456",
			expectedGID: "987654321",
			expectedErr: false,
		},
		{
			name:        "sheet URL with different format",
			url:         "https://docs.google.com/spreadsheets/d/1ABC123DEF456",
			expectedID:  "1ABC123DEF456",
			expectedGID: "0",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, gid, err := utils.ExtractSheetInfo(tt.url)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
				assert.Equal(t, tt.expectedGID, gid)
			}
		})
	}
}

func TestExtractSheetInfo_InvalidURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "invalid URL - no spreadsheet ID",
			url:  "https://docs.google.com/spreadsheets/",
		},
		{
			name: "invalid URL - not a Google Sheets URL",
			url:  "https://example.com/file.csv",
		},
		{
			name: "empty URL",
			url:  "",
		},
		{
			name: "malformed URL",
			url:  "not-a-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := utils.ExtractSheetInfo(tt.url)
			assert.Error(t, err)
		})
	}
}

func TestExtractSheetInfo_EdgeCases(t *testing.T) {
	// Test with very long spreadsheet ID
	longID := "1" + string(make([]byte, 100))
	for i := range longID[1:] {
		longID = longID[:i+1] + "A" + longID[i+2:]
	}

	url := "https://docs.google.com/spreadsheets/d/" + longID + "/edit"
	id, _, err := utils.ExtractSheetInfo(url)
	require.NoError(t, err)
	assert.Equal(t, longID, id)

	// Test with GID that's not numeric (should still work)
	url = "https://docs.google.com/spreadsheets/d/1ABC123/edit#gid=abc123"
	_, gid, err := utils.ExtractSheetInfo(url)
	require.NoError(t, err)
	assert.Equal(t, "abc123", gid)

	// Test with multiple query parameters
	url = "https://docs.google.com/spreadsheets/d/1ABC123/export?format=csv&gid=123&other=value"
	_, gid, err = utils.ExtractSheetInfo(url)
	require.NoError(t, err)
	assert.Equal(t, "123", gid)
}