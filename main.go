package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
)

func main() {
	// Get access token from environment
	accessToken := os.Getenv("ASANA_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Error: ASANA_ACCESS_TOKEN environment variable is not set")
		os.Exit(1)
	}

	// Create Asana client
	client := asana.NewClient(accessToken)

	run(client)
}

func run(client *asana.Client) {
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

	// Print tasks in "My Tasks" by section
	fmt.Println("\nMy Asana Tasks:")
	fmt.Println("==============")

	totalTasks := 0
	sectionsWithTasks := 0

	// Loop through each section and get its tasks
	for _, section := range sections {
		// Get tasks for this section
		tasks, err := client.GetTasksInSection(section.GID)
		if err != nil {
			fmt.Printf("Error getting tasks for section %s: %v\n", section.Name, err)
			continue
		}

		if len(tasks) > 0 {
			fmt.Printf("\n## %s (%d tasks)\n", section.Name, len(tasks))
			sectionsWithTasks++

			for i, task := range tasks {
				// Format the task name with due date if available
				taskLine := fmt.Sprintf("%d. %s", i+1, task.Name)

				// Add due date if present
				if !task.DueOn.IsZero() {
					taskLine = fmt.Sprintf("%s (%s)", taskLine, task.DueOn.Format("2006-01-02"))
				}

				// Print task
				fmt.Println(taskLine)

				// Print notes if present
				if task.Notes != "" {
					// Indent notes
					notes := strings.Split(task.Notes, "\n")
					for _, note := range notes {
						fmt.Printf("     %s\n", note)
					}
				}

				// Add blank line between tasks
				fmt.Println()
			}

			totalTasks += len(tasks)
		}
	}

	if totalTasks == 0 {
		fmt.Println("No tasks found in any section")
	} else {
		fmt.Printf("\nFound %d tasks in %d sections\n", totalTasks, sectionsWithTasks)
	}
}
