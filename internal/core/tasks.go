package core

import (
	"context"
	"fmt"
	"time"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
)

// TaskMove represents a task that should be moved to a new section
type TaskMove struct {
	Task        asana.Task
	SectionGID  string
	SectionName string
}

// CategorizeTasks sorts a list of tasks into categories based on due date
func CategorizeTasks(tasks []asana.Task, now time.Time) map[asana.TaskCategory][]asana.Task {
	categorized := make(map[asana.TaskCategory][]asana.Task)

	for _, task := range tasks {
		category := task.GetTaskCategory(now)
		categorized[category] = append(categorized[category], task)
	}

	return categorized
}

// GetCategoryToSectionMap creates a mapping from task categories to section names based on config
func GetCategoryToSectionMap(config SectionConfig) map[asana.TaskCategory]string {
	return map[asana.TaskCategory]string{
		asana.Overdue:     config.Overdue,
		asana.DueToday:    config.DueToday,
		asana.DueThisWeek: config.DueThisWeek,
		asana.DueLater:    config.DueLater,
		asana.NoDate:      config.NoDate,
	}
}

// CalculateTaskMoves determines which tasks need to be moved to which sections without side effects
// It is a pure function that doesn't perform any side effects
func CalculateTaskMoves(tasks []asana.Task, config SectionConfig,
	sectionNameToGID map[string]string, ignoredSections map[string]bool, now time.Time) []TaskMove {

	var moves []TaskMove

	// Map categories to section names
	categoryToSection := GetCategoryToSectionMap(config)

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

// EnsureRequiredSections creates any missing required sections
func EnsureRequiredSections(ctx context.Context, client asana.API, projectGID string, config SectionConfig,
	sections *[]asana.Section, sectionNameToGID map[string]string) error {

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
			newSection, err := client.CreateSection(ctx, projectGID, sectionName)
			if err != nil {
				return fmt.Errorf("error creating section '%s': %w", sectionName, err)
			}
			sectionNameToGID[newSection.Name] = newSection.GID
			*sections = append(*sections, *newSection)
		}
	}
	return nil
}

// ExecuteTaskMoves performs the actual moves in Asana
func ExecuteTaskMoves(ctx context.Context, client asana.API, taskMoves []TaskMove) error {
	if len(taskMoves) == 0 {
		fmt.Println("\nNo tasks need to be moved")
		return nil
	}

	fmt.Println("\nMoving tasks to appropriate sections...")
	errors := 0

	for _, move := range taskMoves {
		fmt.Printf("Moving task '%s' to section: %s\n", move.Task.Name, move.SectionName)
		err := client.MoveTaskToSection(ctx, move.SectionGID, move.Task.GID)
		if err != nil {
			fmt.Printf("Error moving task '%s': %v\n", move.Task.Name, err)
			errors++
		}
	}

	if errors > 0 {
		return fmt.Errorf("%d errors occurred while moving tasks", errors)
	}

	fmt.Printf("\nMoved %d tasks to their appropriate sections\n", len(taskMoves))
	return nil
}

// CreateIgnoredSectionsMap converts a slice of ignored section names to a map for quick lookup
func CreateIgnoredSectionsMap(ignoredSections []string) map[string]bool {
	result := make(map[string]bool)
	for _, sectionName := range ignoredSections {
		result[sectionName] = true
	}
	return result
}

// CreateSectionNameToGIDMap creates a mapping of section names to their GIDs
func CreateSectionNameToGIDMap(sections []asana.Section) map[string]string {
	result := make(map[string]string)
	for _, section := range sections {
		result[section.Name] = section.GID
	}
	return result
}

// OrganizeTasks is the main business logic function that fetches and organizes tasks
func OrganizeTasks(ctx context.Context, client asana.API, config SectionConfig, dryRun bool) (map[asana.TaskCategory][]asana.Task, error) {
	// Get current user
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}
	fmt.Printf("Logged in as: %s\n", user.Name)

	// Get workspaces
	workspaces, err := client.GetWorkspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		return nil, fmt.Errorf("no workspaces found for user")
	}

	// Use first workspace
	workspace := workspaces[0]
	fmt.Printf("Using workspace: %s\n", workspace.Name)

	// Get user's "My Tasks" list
	userTaskList, err := client.GetUserTaskList(ctx, user.GID, workspace.GID)
	if err != nil {
		return nil, fmt.Errorf("error getting user task list: %w", err)
	}

	// Get sections in My Tasks list (using the project/sections API)
	sections, err := client.GetSectionsForProject(ctx, userTaskList.GID)
	if err != nil {
		return nil, fmt.Errorf("error getting sections: %w", err)
	}

	// Create a map to store section names to their GIDs
	sectionNameToGID := CreateSectionNameToGIDMap(sections)

	// Ensure required sections exist, create them if needed
	if !dryRun {
		if err := EnsureRequiredSections(ctx, client, userTaskList.GID, config, &sections, sectionNameToGID); err != nil {
			return nil, fmt.Errorf("error ensuring required sections: %w", err)
		}
	}

	// Create a map of ignored sections for quick lookup
	ignoredSections := CreateIgnoredSectionsMap(config.IgnoredSections)

	// Collect all tasks from user task list at once
	fmt.Println("Fetching all tasks from My Tasks list...")
	allTasks, err := client.GetTasksFromUserTaskList(ctx, userTaskList.GID)
	if err != nil {
		return nil, fmt.Errorf("error getting tasks from user task list: %w", err)
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
	categorizedTasks := CategorizeTasks(allTasks, now)

	// Calculate task moves without side effects
	taskMoves := CalculateTaskMoves(allTasks, config, sectionNameToGID, ignoredSections, now)

	// Execute the moves if not in dry run mode
	if !dryRun && len(taskMoves) > 0 {
		if err := ExecuteTaskMoves(ctx, client, taskMoves); err != nil {
			return nil, fmt.Errorf("error executing task moves: %w", err)
		}
	}

	return categorizedTasks, nil
}
