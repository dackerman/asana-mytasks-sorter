package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
	"github.com/dackerman/asana-tasks-sorter/internal/config"
	"github.com/dackerman/asana-tasks-sorter/internal/core"
	"github.com/dackerman/asana-tasks-sorter/internal/ui"
)

func main() {
	// Parse command-line flags
	configFile := flag.String("config", "", "Path to section configuration file")
	dryRun := flag.Bool("dry-run", false, "Only display changes without moving tasks")
	flag.Parse()

	// Get access token from environment
	accessToken := os.Getenv("ASANA_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Error: ASANA_ACCESS_TOKEN environment variable is not set")
		os.Exit(1)
	}

	// Create Asana client
	client := asana.NewClient(accessToken)

	// Load configuration
	config := config.LoadConfiguration(*configFile)

	// Run the main business logic
	categorizedTasks, err := core.OrganizeTasks(client, config, *dryRun)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Display the tasks in a formatted way
	ui.DisplayTasks(categorizedTasks, core.GetCategoryToSectionMap(config), *dryRun)
}