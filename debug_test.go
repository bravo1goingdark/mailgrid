package main

import (
	"fmt"
	"strings"

	"github.com/bravo1goingdark/mailgrid/parser"
)

func main() {
	csvData := "email,name,company\nuser1@example.com,Alice,Acme\nuser2@example.com,Bob,Widgets\n"
	r := strings.NewReader(csvData)

	recipients, err := parser.ParseCSVFromReader(r)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Number of recipients: %d\n", len(recipients))
	for i, recipient := range recipients {
		fmt.Printf("Recipient %d:\n", i)
		fmt.Printf("  Email: %s\n", recipient.Email)
		fmt.Printf("  Data: %+v\n", recipient.Data)
	}
}
