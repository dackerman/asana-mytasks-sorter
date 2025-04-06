package ui

import (
	"fmt"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
)

// DisplayTasks prints out tasks organized by category with color formatting
func DisplayTasks(categorizedTasks map[asana.TaskCategory][]asana.Task,
	categoryToSection map[asana.TaskCategory]string, dryRun bool) {

	fmt.Printf("\n%s\n", Header("My Asana Tasks:"))
	fmt.Println(Bold + "==============" + Reset)

	totalTasks := 0
	sectionsWithTasks := 0

	// Print tasks by category
	for category, sectionName := range categoryToSection {
		tasks := categorizedTasks[category]
		if len(tasks) > 0 {
			// Format section header with color based on category
			var sectionHeader string
			var taskColor string

			switch category {
			case asana.Overdue:
				taskColor = BrightRed
			case asana.DueToday:
				taskColor = BrightYellow
			case asana.DueThisWeek:
				taskColor = BrightGreen
			default:
				taskColor = BrightCyan
			}
			sectionHeader = fmt.Sprintf("\n%s %s (%d tasks)\n", Bold+taskColor+"##", sectionName+Reset, len(tasks))
			fmt.Print(sectionHeader + Reset)
			sectionsWithTasks++

			for i, task := range tasks {
				// Format the task number
				taskNum := fmt.Sprintf("%d. ", i+1)

				// Format the task name matching section color
				taskNameStr := TaskName(task.Name)

				// Format the due date if present
				var dueStr string
				if !task.DueOn.IsZero() {
					dueStr = " " + DueDate("("+task.DueOn.Format("2006-01-02")+")")
				}

				fmt.Println(taskNum + taskNameStr + dueStr)
			}

			totalTasks += len(tasks)
		}
	}

	if totalTasks == 0 {
		fmt.Println(Warning("No tasks found in any section"))
	} else {
		fmt.Printf("\n%s\n", Success(fmt.Sprintf("Found %d tasks in %d sections", totalTasks, sectionsWithTasks)))
	}

	if dryRun {
		fmt.Println("\n" + Important(Warning("This was a dry run. To actually move tasks, run without the --dry-run flag.")))
	}
}

// FatalError prints an error message and exits the program
func FatalError(format string, args ...interface{}) {
	errorMsg := fmt.Sprintf(format, args...)
	fmt.Println(Error("ERROR: " + errorMsg))
	panic(errorMsg)
}
