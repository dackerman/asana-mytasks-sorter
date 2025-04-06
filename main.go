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

	// Load configuration
	config := loadConfiguration(*configFile)

	// Run the main logic
	run(client, config, *dryRun)
}

// loadConfiguration loads the configuration from a file or returns defaults
func loadConfiguration(configFile string) asana.SectionConfig {
	config := asana.DefaultSectionConfig()
	
	if configFile == "" {
		return config
	}
	
	loadedConfig, err := loadSectionConfig(configFile)
	if err == nil {
		return loadedConfig
	}
	
	fmt.Printf("Error loading section config: %v\nUsing default configuration\n", err)
	return config
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

// TaskMove represents a task that should be moved to a new section
type TaskMove struct {
	Task        asana.Task
	SectionGID  string
	SectionName string
}

// categorizeTasks sorts a list of tasks into categories based on due date
func categorizeTasks(tasks []asana.Task, now time.Time) map[asana.TaskCategory][]asana.Task {
	categorized := make(map[asana.TaskCategory][]asana.Task)
	
	for _, task := range tasks {
		category := task.GetTaskCategory(now)
		categorized[category] = append(categorized[category], task)
	}
	
	return categorized
}

// getCategoryToSectionMap creates a mapping from task categories to section names based on config
func getCategoryToSectionMap(config asana.SectionConfig) map[asana.TaskCategory]string {
	return map[asana.TaskCategory]string{
		asana.Overdue:     config.Overdue,
		asana.DueToday:    config.DueToday,
		asana.DueThisWeek: config.DueThisWeek,
		asana.DueLater:    config.DueLater,
		asana.NoDate:      config.NoDate,
	}
}

// calculateTaskMoves determines which tasks need to be moved to which sections without side effects
// It is a pure function that doesn't perform any side effects
func calculateTaskMoves(tasks []asana.Task, config asana.SectionConfig, 
	sectionNameToGID map[string]string, ignoredSections map[string]bool, now time.Time) []TaskMove {
	
	var moves []TaskMove
	
	// Map categories to section names
	categoryToSection := getCategoryToSectionMap(config)
	
	for _, task := range tasks {
		// Get current section name
		currentSectionName := task.AssigneeSection.Name
		
		// Skip if task is in an ignored section
		if ignoredSections[currentSectionName] {
			continue
		}
		
		// Calculate which category the task belongs in
		category := task.GetTaskCategory(now)
		
		// Get the target section name for this category
		targetSectionName := categoryToSection[category]
		
		// Skip if target section is in the ignored list
		if ignoredSections[targetSectionName] {
			continue
		}
		
		// Skip if task is already in the correct section
		if currentSectionName == targetSectionName {
			continue
		}
		
		// Get the section GID for the target section
		sectionGID, exists := sectionNameToGID[targetSectionName]
		if !exists {
			continue
		}
		
		// Add the move to our list
		moves = append(moves, TaskMove{
			Task:        task,
			SectionGID:  sectionGID,
			SectionName: targetSectionName,
		})
	}
	
	return moves
}

// fatalError prints an error message and exits the program
func fatalError(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}

func run(client *asana.Client, config asana.SectionConfig, dryRun bool) {
	// Get current user
	user, err := client.GetCurrentUser()
	if err != nil {
		fatalError("Error getting current user: %v\n", err)
	}
	fmt.Printf("Logged in as: %s\n", user.Name)

	// Get workspaces
	workspaces, err := client.GetWorkspaces()
	if err != nil {
		fatalError("Error getting workspaces: %v\n", err)
	}

	if len(workspaces) == 0 {
		fatalError("No workspaces found for user\n")
	}

	// Use first workspace
	workspace := workspaces[0]
	fmt.Printf("Using workspace: %s\n", workspace.Name)

	// Get user's "My Tasks" list
	userTaskList, err := client.GetUserTaskList(user.GID, workspace.GID)
	if err != nil {
		fatalError("Error getting user task list: %v\n", err)
	}

	// Get sections in My Tasks list (using the project/sections API)
	sections, err := client.GetSectionsForProject(userTaskList.GID)
	if err != nil {
		fatalError("Error getting sections: %v\n", err)
	}

	// Create a map to store section names to their GIDs
	sectionNameToGID := make(map[string]string)
	for _, section := range sections {
		sectionNameToGID[section.Name] = section.GID
	}

	// Ensure required sections exist, create them if needed
	if !dryRun {
		ensureRequiredSections(client, userTaskList.GID, config, &sections, sectionNameToGID)
	}

	// Create a map of ignored sections for quick lookup
	ignoredSections := make(map[string]bool)
	for _, sectionName := range config.IgnoredSections {
		ignoredSections[sectionName] = true
	}

	// Collect all tasks from user task list at once
	fmt.Println("Fetching all tasks from My Tasks list...")
	allTasks, err := client.GetTasksFromUserTaskList(userTaskList.GID)
	if err != nil {
		fatalError("Error getting tasks from user task list: %v\n", err)
	}

	// Print tasks we're skipping due to being in ignored sections
	for _, task := range allTasks {
		sectionName := task.AssigneeSection.Name
		if ignoredSections[sectionName] {
			fmt.Printf("Skipping task in ignored section: %s (in section '%s')\n",
				task.Name, sectionName)
		}
	}

	// Get the current time once for consistency across all operations
	now := time.Now()

	// Sort tasks into categories based on due date for display purposes
	categorizedTasks := categorizeTasks(allTasks, now)

	// Map categories to section names (used for display)
	categoryToSection := getCategoryToSectionMap(config)

	// Calculate task moves without side effects
	taskMoves := calculateTaskMoves(allTasks, config, sectionNameToGID, ignoredSections, now)
	
	// Execute the moves if not in dry run mode
	if !dryRun {
		executeTaskMoves(client, taskMoves)
	}

	// Display tasks by category
	displayTasks(categorizedTasks, categoryToSection, dryRun)
}

// executeTaskMoves performs the actual moves in Asana
func executeTaskMoves(client *asana.Client, taskMoves []TaskMove) {
	if len(taskMoves) == 0 {
		fmt.Println("\nNo tasks need to be moved")
		return
	}
	
	fmt.Println("\nMoving tasks to appropriate sections...")
	tasksMoved := 0
	
	for _, move := range taskMoves {
		fmt.Printf("Moving task '%s' to section: %s\n", move.Task.Name, move.SectionName)
		err := client.MoveTaskToSection(move.SectionGID, move.Task.GID)
		if err != nil {
			fmt.Printf("Error moving task '%s': %v\n", move.Task.Name, err)
		} else {
			tasksMoved++
		}
	}
	
	fmt.Printf("\nMoved %d tasks to their appropriate sections\n", tasksMoved)
}

// ensureRequiredSections creates any missing required sections
func ensureRequiredSections(client *asana.Client, projectGID string, config asana.SectionConfig,
	sections *[]asana.Section, sectionNameToGID map[string]string) {
	
	// List of required sections from config
	requiredSections := []string{
		config.Overdue,
		config.DueToday,
		config.DueThisWeek,
		config.DueLater,
		config.NoDate,
	}
	
	for _, sectionName := range requiredSections {
		if _, exists := sectionNameToGID[sectionName]; !exists {
			fmt.Printf("Creating section: %s\n", sectionName)
			newSection, err := client.CreateSection(projectGID, sectionName)
			if err != nil {
				fmt.Printf("Error creating section '%s': %v\n", sectionName, err)
				continue
			}
			sectionNameToGID[newSection.Name] = newSection.GID
			*sections = append(*sections, *newSection)
		}
	}
}

// displayTasks prints out tasks organized by category
func displayTasks(categorizedTasks map[asana.TaskCategory][]asana.Task, 
	categoryToSection map[asana.TaskCategory]string, dryRun bool) {
	
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
