// cmd/mailgrid/main.go
package main

import (
	"log"
	"mailgrid/cli"
)

// main is the CLI entry point for mailgrid.
// It parses CLI flags and delegates execution to the CLI runner.
func main() {
	// Parse CLI flags into a structured config
	args := cli.ParseFlags()

	// Run the mailgrid workflow (load config, parse CSV, render/send emails)
	if err := cli.Run(args); err != nil {
		log.Fatalf("\u274c %v", err)
	}
}
