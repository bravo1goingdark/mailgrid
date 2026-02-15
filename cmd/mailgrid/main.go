// cmd/mailgrid/main.go
package main

import (
	"fmt"
	"github.com/bravo1goingdark/mailgrid/cli"
	"log"
	"os"
)

// Version information (set at build time)
var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

// main is the CLI entry point for mailgrid.
// It parses CLI flags and delegates execution to the CLI runner.
func main() {
	// Parse CLI flags into a structured config
	args := cli.ParseFlags()

	// Handle help flag (help shown by ParseFlags, just exit)
	if args.ShowHelp {
		return
	}

	// Handle version flag
	if args.ShowVersion {
		showVersion()
		return
	}

	// Run the mailgrid workflow (load config, parse CSV, render/send emails)
	if err := cli.Run(args); err != nil {
		log.Fatalf("❌ %v", err)
	}
}

// showVersion displays version information
func showVersion() {
	fmt.Printf("MailGrid v%s\n", version)
	fmt.Printf("Built: %s\n", buildTime)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("\nMailGrid is a production-ready email orchestrator for bulk email campaigns.\n")
	fmt.Printf("Documentation: https://github.com/bravo1goingdark/mailgrid\n")
	os.Exit(0)
}
