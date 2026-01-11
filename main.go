package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tom/shippost/config"
	"github.com/tom/shippost/tui"
	"github.com/tom/shippost/x"
)

const version = "0.2.0"

func main() {
	// Define flags
	setup := flag.Bool("setup", false, "Configure X API credentials")
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

	if *setup {
		if err := config.RunSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Get post text from arguments
	args := flag.Args()

	// No arguments - launch TUI
	if len(args) == 0 {
		if err := tui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Quick mode - post directly
	text := strings.Join(args, " ")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create client and post
	client := x.NewClient(cfg)
	resp, err := client.Post(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Posted: %s\n", resp.Data.Text)
	fmt.Printf("https://x.com/i/status/%s\n", resp.Data.ID)
}

func printUsage() {
	fmt.Println("shippost - Post to X from your terminal")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  shippost                           Launch TUI (browse commits, compose)")
	fmt.Println("  shippost \"Your post text here\"     Quick post to X")
	fmt.Println("  shippost --setup                   Configure X API credentials")
	fmt.Println("  shippost --version                 Show version")
	fmt.Println("  shippost --help                    Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  shippost                           Interactive mode")
	fmt.Println("  shippost \"Just shipped a new feature!\"")
}
