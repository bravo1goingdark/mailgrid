package parser

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net/mail"
	"os"
	"strings"
)

// ValidateEmail checks if an email address is valid using net/mail.
// Returns nil if valid, error otherwise.
func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

// IsValidEmail returns true if the email address is valid.
func IsValidEmail(email string) bool {
	return ValidateEmail(email) == nil
}

// ParseCSVFromReader reads a CSV from any io.Reader and returns a list of Recipients.
// It expects one column to be named 'email' and uses other columns as dynamic data.
// Invalid email addresses are skipped with a warning.
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
	var skippedCount int
	var totalRows int

	// Read the remaining rows (one recipient per row)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break // end of file reached
		}
		totalRows++
		if err != nil || len(record) != len(headers) {
			skippedCount++
			log.Printf("Warning: Skipping malformed CSV row (expected %d fields, got %d)", len(headers), len(record))
			continue // skip malformed or mismatched rows
		}

		// Grab and clean the email value
		email := strings.TrimSpace(record[emailIdx])
		if email == "" {
			continue // skip blank emails
		}

		// Validate email address
		if !IsValidEmail(email) {
			skippedCount++
			log.Printf("Warning: Skipping row %d with invalid email: %s", totalRows, email)
			continue
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

	// Deduplicate by email (case-insensitive). A CSV with duplicate addresses
	// would otherwise send the same email multiple times.
	seen := make(map[string]struct{}, len(recipients))
	deduped := recipients[:0]
	var dupCount int
	for _, r := range recipients {
		key := strings.ToLower(r.Email)
		if _, exists := seen[key]; exists {
			dupCount++
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, r)
	}
	recipients = deduped

	if skippedCount > 0 || dupCount > 0 {
		log.Printf("CSV parsing: %d recipients loaded, %d rows skipped, %d duplicates removed (total rows: %d)",
			len(recipients), skippedCount, dupCount, totalRows)
	}

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
			log.Printf("Warning: Failed to close CSV file: %v", err)
		}
	}(file)

	// Delegate to reader-based parser
	return ParseCSVFromReader(file)
}
