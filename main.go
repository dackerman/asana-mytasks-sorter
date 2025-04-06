package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
)

func main() {
	// Define command-line flags
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

	// Load section config
	config := asana.DefaultSectionConfig()
	if *configFile != "" {
		loadedConfig, err := loadSectionConfig(*configFile)
		if err == nil {
			config = loadedConfig
		} else {
			fmt.Printf("Error loading section config: %v\nUsing default configuration\n", err)
		}
	}

	run(client, config, *dryRun)
}

// loadSectionConfig loads the section configuration from a JSON file
func loadSectionConfig(configPath string) (asana.SectionConfig, error) {
	// Handle relative paths
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return asana.SectionConfig{}, fmt.Errorf("failed to resolve absolute path: %v", err)
		}
		configPath = absPath
	}

	// Read config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return asana.SectionConfig{}, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse JSON
	var config asana.SectionConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return asana.SectionConfig{}, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

func run(client *asana.Client, config asana.SectionConfig, dryRun bool) {
	// Get current user
	user, err := client.GetCurrentUser()
	if err != nil {
		fmt.Printf("Error getting current user: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Logged in as: %s\n", user.Name)

	// Get workspaces
	workspaces, err := client.GetWorkspaces()
	if err != nil {
		fmt.Printf("Error getting workspaces: %v\n", err)
		os.Exit(1)
	}

	if len(workspaces) == 0 {
		fmt.Println("No workspaces found for user")
		os.Exit(1)
	}

	// Use first workspace
	workspace := workspaces[0]
	fmt.Printf("Using workspace: %s\n", workspace.Name)

	// Get user's "My Tasks" list
	userTaskList, err := client.GetUserTaskList(user.GID, workspace.GID)
	if err != nil {
		fmt.Printf("Error getting user task list: %v\n", err)
		os.Exit(1)
	}

	// Get sections in My Tasks list (using the project/sections API)
	sections, err := client.GetSectionsForProject(userTaskList.GID)
	if err != nil {
		fmt.Printf("Error getting sections: %v\n", err)
		os.Exit(1)
	}

	// Create a map to store section names to their GIDs
	sectionNameToGID := make(map[string]string)
	for _, section := range sections {
		sectionNameToGID[section.Name] = section.GID
	}

	// Ensure required sections exist, create them if needed
	requiredSections := []string{
		config.Overdue,
		config.DueToday,
		config.DueThisWeek,
		config.DueLater,
		config.NoDate,
	}

	if !dryRun {
		for _, sectionName := range requiredSections {
			if _, exists := sectionNameToGID[sectionName]; !exists {
				fmt.Printf("Creating section: %s\n", sectionName)
				newSection, err := client.CreateSection(userTaskList.GID, sectionName)
				if err != nil {
					fmt.Printf("Error creating section '%s': %v\n", sectionName, err)
					continue
				}
				sectionNameToGID[newSection.Name] = newSection.GID
				sections = append(sections, *newSection)
			}
		}
	}

	// Create a map of ignored sections for quick lookup
	ignoredSections := make(map[string]bool)
	for _, sectionName := range config.IgnoredSections {
		ignoredSections[sectionName] = true
	}

	// Create a map of section IDs to names for easy lookup
	sectionGIDToName := make(map[string]string)
	for _, section := range sections {
		sectionGIDToName[section.GID] = section.Name
	}

	// Collect all tasks from user task list at once
	fmt.Println("Fetching all tasks from My Tasks list...")
	allTasks, err := client.GetTasksFromUserTaskList(userTaskList.GID)
	if err != nil {
		fmt.Printf("Error getting tasks from user task list: %v\n", err)
		os.Exit(1)
	}

	// Filter out tasks from ignored sections
	filteredTasks := []asana.Task{}
	for _, task := range allTasks {
		// Skip tasks in ignored sections
		sectionName := task.AssigneeSection.Name
		if ignoredSections[sectionName] {
			fmt.Printf("Skipping task in ignored section: %s (in section '%s')\n",
				task.Name, sectionName)
			continue
		}
		filteredTasks = append(filteredTasks, task)
	}
	allTasks = filteredTasks

	// Sort tasks into categories based on due date
	categorizedTasks := make(map[asana.TaskCategory][]asana.Task)
	now := time.Now()

	for _, task := range allTasks {
		category := task.GetTaskCategory(now)
		categorizedTasks[category] = append(categorizedTasks[category], task)
	}

	// Map categories to section names
	categoryToSection := map[asana.TaskCategory]string{
		asana.Overdue:     config.Overdue,
		asana.DueToday:    config.DueToday,
		asana.DueThisWeek: config.DueThisWeek,
		asana.DueLater:    config.DueLater,
		asana.NoDate:      config.NoDate,
	}

	// Move tasks to appropriate sections
	tasksMoved := 0
	if !dryRun {
		fmt.Println("\nMoving tasks to appropriate sections...")
		for category, sectionName := range categoryToSection {
			// Skip if target section is in the ignored list
			if ignoredSections[sectionName] {
				fmt.Printf("Skipping moving tasks to ignored section: %s\n", sectionName)
				continue
			}

			sectionGID, exists := sectionNameToGID[sectionName]
			if !exists {
				fmt.Printf("Error: Section '%s' not found, skipping tasks\n", sectionName)
				continue
			}

			tasks := categorizedTasks[category]
			for _, task := range tasks {
				// Get current section name directly from the task
				currentSectionName := task.AssigneeSection.Name

				// Skip if task is already in the correct section
				if currentSectionName == sectionName {
					fmt.Printf("Task '%s' already in correct section: %s\n", task.Name, sectionName)
					continue
				}

				// Skip if task is currently in an ignored section
				if ignoredSections[currentSectionName] {
					fmt.Printf("Task '%s' is in ignored section '%s', skipping\n", task.Name, currentSectionName)
					continue
				}

				// Move task to the new section
				fmt.Printf("Moving task '%s' to section: %s\n", task.Name, sectionName)
				err := client.MoveTaskToSection(sectionGID, task.GID)
				if err != nil {
					fmt.Printf("Error moving task '%s': %v\n", task.Name, err)
				} else {
					tasksMoved++
				}
			}
		}
		fmt.Printf("\nMoved %d tasks to their appropriate sections\n", tasksMoved)
	}

	// Print tasks by category in a concise format
	fmt.Println("\nMy Asana Tasks:")
	fmt.Println("==============")

	totalTasks := 0
	sectionsWithTasks := 0

	// Print tasks by category
	for category, sectionName := range categoryToSection {
		tasks := categorizedTasks[category]
		if len(tasks) > 0 {
			fmt.Printf("\n## %s (%d tasks)\n", sectionName, len(tasks))
			sectionsWithTasks++

			for i, task := range tasks {
				// Format the task name with due date if available
				taskLine := fmt.Sprintf("%d. %s", i+1, task.Name)

				// Add due date if present
				if !task.DueOn.IsZero() {
					taskLine = fmt.Sprintf("%s (%s)", taskLine, task.DueOn.Format("2006-01-02"))
				}

				fmt.Println(taskLine)
			}

			totalTasks += len(tasks)
		}
	}

	if totalTasks == 0 {
		fmt.Println("No tasks found in any section")
	} else {
		fmt.Printf("\nFound %d tasks in %d sections\n", totalTasks, sectionsWithTasks)
	}

	if dryRun {
		fmt.Println("\nThis was a dry run. To actually move tasks, run without the --dry-run flag.")
	}
}
