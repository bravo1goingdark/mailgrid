package parser

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

// ParseCSVFromReader reads a CSV from any io.Reader and returns a list of Recipients.
// It expects one column to be named 'email' and uses other columns as dynamic data.
func ParseCSVFromReader(reader io.Reader) ([]Recipient, error) {
	// create a new CSV reader instance
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true // clean up any accidental spaces

	// read the first row as header (column names)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	// normalize headers to lowercase and trim extra spaces
	for i, h := range headers {
		headers[i] = strings.ToLower(strings.TrimSpace(h))
	}

	emailIdx := -1
	for i, h := range headers {
		if h == "email" {
			emailIdx = i
			break
		}
	}
	if emailIdx == -1 {
		return nil, errors.New("CSV must include 'email' column")
	}

	var recipients []Recipient

	// Read the remaining rows (one recipient per row)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break // end of file reached
		}
		if err != nil || len(record) != len(headers) {
			continue // skip malformed or mismatched rows
		}

		// Grab and clean the email value
		email := strings.TrimSpace(record[emailIdx])
		if email == "" {
			continue // skip blank emails
		}

		// Collect all the other data fields (except email)
		data := make(map[string]string)
		for i, value := range record {
			key := headers[i]
			if key == "email" {
				continue
			}
			data[key] = strings.TrimSpace(value)
		}

		// Add the parsed recipient to our result list
		recipients = append(recipients, Recipient{
			Email: email,
			Data:  data,
		})
	}

	// Return the full list of recipients
	return recipients, nil
}

// ParseCSV reads a CSV file from a given path and returns a list of Recipients.
// This is a wrapper around ParseCSVFromReader for convenience.
func ParseCSV(path string) ([]Recipient, error) {
	// Open the CSV file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// close the file when done (even if an error occurs later)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Failed to close config file: %v", err)
		}
	}(file)

	// Delegate to reader-based parser
	return ParseCSVFromReader(file)
}
