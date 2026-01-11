package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tomswokowski/shippost/config"
	"github.com/tomswokowski/shippost/tui"
)

// version is set by goreleaser at build time
var version = "dev"

func main() {
	// Define flags
	setup := flag.Bool("setup", false, "Configure X API credentials")
	cleanup := flag.Bool("cleanup", false, "Remove stored credentials")
	showVersion := flag.Bool("version", false, "Show version")
	help := flag.Bool("help", false, "Show help")

	flag.Usage = printUsage
	flag.Parse()

	// Handle flags
	if *showVersion {
		fmt.Printf("shippost v%s\n", version)
		return
	}

	if *help {
		printUsage()
		return
	}

	if *cleanup {
		if err := config.Cleanup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *setup {
		if err := config.RunSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Auto-run setup if no config exists
	if !config.Exists() {
		if err := config.RunSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println()
	}

	// Launch TUI
	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("shippost - Post to X from your terminal")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  shippost            Launch the app")
	fmt.Println("  shippost --setup    Configure X API credentials")
	fmt.Println("  shippost --cleanup  Remove stored credentials")
	fmt.Println("  shippost --version  Show version")
	fmt.Println("  shippost --help     Show this help")
}
