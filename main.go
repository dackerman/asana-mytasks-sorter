package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
	"github.com/dackerman/asana-tasks-sorter/internal/config"
	"github.com/dackerman/asana-tasks-sorter/internal/core"
	"github.com/dackerman/asana-tasks-sorter/internal/ui"
)

const usageText = `Asana Tasks Sorter - A CLI tool to organize your Asana tasks.

This program connects to your Asana account, fetches your tasks, and 
organizes them into sections based on their due dates. It can optionally
move the tasks to the appropriate sections in Asana.

Usage:
  asana-tasks-sorter [flags]

Configuration:
  Create a JSON file with the following structure to customize section names:
  {
    "overdue": "Overdue",
    "due_today": "Due today",
    "due_this_week": "Due within the next 7 days",
    "due_later": "Due later",
    "no_date": "Recently assigned",
    "ignored_sections": ["Doing Now", "Waiting For"]
  }

Authentication:
  Export your Asana personal access token as an environment variable:
  export ASANA_ACCESS_TOKEN="your_asana_personal_access_token"
  Get your token from https://app.asana.com/0/developer-console

Examples:
  # Run with default settings
  asana-tasks-sorter

  # Use a custom configuration file
  asana-tasks-sorter -config path/to/your/config.json

  # Preview changes without moving tasks
  asana-tasks-sorter -dry-run

  # Set a custom timeout for API operations
  asana-tasks-sorter -timeout 60s
`

func main() {
	// Set custom usage text
	flag.Usage = func() {
		fmt.Println(ui.Header("Asana Tasks Sorter - A CLI tool to organize your Asana tasks."))
		fmt.Println()

		description := `This program connects to your Asana account, fetches your tasks, and 
organizes them into sections based on their due dates. It can optionally
move the tasks to the appropriate sections in Asana.`
		fmt.Println(description)
		fmt.Println()

		fmt.Println(ui.SectionTitle("Usage:"))
		fmt.Println("  asana-tasks-sorter [flags]")
		fmt.Println()

		fmt.Println(ui.SectionTitle("Configuration:"))
		configText := `  Create a JSON file with the following structure to customize section names:
  {
    "overdue": "Overdue",
    "due_today": "Due today",
    "due_this_week": "Due within the next 7 days",
    "due_later": "Due later",
    "no_date": "Recently assigned",
    "ignored_sections": ["Doing Now", "Waiting For"]
  }`
		fmt.Println(configText)
		fmt.Println()

		fmt.Println(ui.SectionTitle("Authentication:"))
		authText := `  Export your Asana personal access token as an environment variable:
  export ASANA_ACCESS_TOKEN="your_asana_personal_access_token"
  Get your token from https://app.asana.com/0/developer-console`
		fmt.Println(authText)
		fmt.Println()

		fmt.Println(ui.SectionTitle("Examples:"))
		examplesText := `  # Run with default settings
  asana-tasks-sorter --config default

  # Use a custom configuration file
  asana-tasks-sorter --config path/to/your/config.json

  # Preview changes without moving tasks
  asana-tasks-sorter --config default --dry-run

  # Set a custom timeout for API operations
  asana-tasks-sorter --config default --timeout 60s`
		fmt.Println(examplesText)
		fmt.Println()

		fmt.Println(ui.SectionTitle("Flags:"))
		flag.PrintDefaults()
	}

	// Parse command-line flags
	configFile := flag.String("config", "", "Path to section configuration file or 'default' to use built-in defaults (required)")
	dryRun := flag.Bool("dry-run", false, "Only display changes without moving tasks")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout for API operations")
	help := flag.Bool("help", false, "Show detailed help information")
	flag.Parse()

	// Show help if requested or if no config is provided
	if *help || flag.NArg() > 0 || *configFile == "" {
		flag.Usage()
		if !*help && *configFile == "" {
			fmt.Println("\n" + ui.Error("Error: --config parameter is required"))
			fmt.Println(ui.Info("Use --config default to use the built-in defaults"))
		}
		return
	}

	// Get access token from environment
	accessToken := os.Getenv("ASANA_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Error: ASANA_ACCESS_TOKEN environment variable is not set")
		os.Exit(1)
	}

	// Create Asana client
	client := asana.NewClient(accessToken)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Load configuration
	conf := config.LoadConfiguration(*configFile)

	// Run the main business logic
	categorizedTasks, err := core.OrganizeTasks(ctx, client, conf, *dryRun)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Display the tasks in a formatted way
	ui.DisplayTasks(categorizedTasks, core.GetCategoryToSectionMap(conf), *dryRun)
}
