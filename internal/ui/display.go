package ui

import (
	"fmt"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
)

// DisplayTasks prints out tasks organized by category
func DisplayTasks(categorizedTasks map[asana.TaskCategory][]asana.Task, 
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

// FatalError prints an error message and exits the program
func FatalError(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	panic(fmt.Sprintf(format, args...))
}