package main

import (
	"testing"
	"time"

	"github.com/dackerman/asana-tasks-sorter/internal/asana"
	"github.com/dackerman/asana-tasks-sorter/internal/core"
)

func TestCalculateTaskMoves(t *testing.T) {
	// Define a fixed reference time for tests (2023-04-15)
	referenceTime := time.Date(2023, 4, 15, 12, 0, 0, 0, time.UTC)

	// Helper function to create a test task
	createTask := func(name string, dueDate string, sectionName string) asana.Task {
		var parsedDate asana.Date
		if dueDate != "" {
			t, err := time.Parse("2006-01-02", dueDate)
			if err != nil {
				panic(err)
			}
			parsedDate = asana.Date(t)
		}

		return asana.Task{
			GID:   "task_" + name,
			Name:  name,
			DueOn: parsedDate,
			AssigneeSection: asana.AssigneeSection{
				GID:  "section_" + sectionName,
				Name: sectionName,
			},
		}
	}

	// Test cases
	testCases := []struct {
		name             string
		tasks            []asana.Task
		config           core.SectionConfig
		sectionNameToGID map[string]string
		ignoredSections  map[string]bool
		expectedMoves    []core.TaskMove
	}{
		{
			name: "Tasks in correct sections should not move",
			tasks: []asana.Task{
				createTask("Task 1", "2023-04-15", "Due today"),                  // Today's date, already in Due today
				createTask("Task 2", "2023-04-10", "Overdue"),                    // Past date, already in Overdue
				createTask("Task 3", "2023-04-20", "Due within the next 7 days"), // Within a week, already correct
				createTask("Task 4", "2023-05-15", "Due later"),                  // More than a week, already correct
				createTask("Task 5", "", "Recently assigned"),                    // No date, already in correct section
			},
			config: core.SectionConfig{
				Overdue:     "Overdue",
				DueToday:    "Due today",
				DueThisWeek: "Due within the next 7 days",
				DueLater:    "Due later",
				NoDate:      "Recently assigned",
			},
			sectionNameToGID: map[string]string{
				"Overdue":                    "section_Overdue",
				"Due today":                  "section_Due today",
				"Due within the next 7 days": "section_Due within the next 7 days",
				"Due later":                  "section_Due later",
				"Recently assigned":          "section_Recently assigned",
			},
			ignoredSections: map[string]bool{},
			expectedMoves:   []core.TaskMove{}, // No moves expected
		},
		{
			name: "Tasks in wrong sections should move",
			tasks: []asana.Task{
				createTask("Task 1", "2023-04-15", "Overdue"),                    // Today's date, should move to Due today
				createTask("Task 2", "2023-04-10", "Due today"),                  // Past date, should move to Overdue
				createTask("Task 3", "2023-04-20", "Due later"),                  // Within a week, should move to Due this week
				createTask("Task 4", "2023-05-15", "Due within the next 7 days"), // Far future, should move to Due later
				createTask("Task 5", "", "Due today"),                            // No date, should move to Recently assigned
			},
			config: core.SectionConfig{
				Overdue:     "Overdue",
				DueToday:    "Due today",
				DueThisWeek: "Due within the next 7 days",
				DueLater:    "Due later",
				NoDate:      "Recently assigned",
			},
			sectionNameToGID: map[string]string{
				"Overdue":                    "section_Overdue",
				"Due today":                  "section_Due today",
				"Due within the next 7 days": "section_Due within the next 7 days",
				"Due later":                  "section_Due later",
				"Recently assigned":          "section_Recently assigned",
			},
			ignoredSections: map[string]bool{},
			expectedMoves: []core.TaskMove{
				{
					Task:        createTask("Task 1", "2023-04-15", "Overdue"),
					SectionGID:  "section_Due today",
					SectionName: "Due today",
				},
				{
					Task:        createTask("Task 2", "2023-04-10", "Due today"),
					SectionGID:  "section_Overdue",
					SectionName: "Overdue",
				},
				{
					Task:        createTask("Task 3", "2023-04-20", "Due later"),
					SectionGID:  "section_Due within the next 7 days",
					SectionName: "Due within the next 7 days",
				},
				{
					Task:        createTask("Task 4", "2023-05-15", "Due within the next 7 days"),
					SectionGID:  "section_Due later",
					SectionName: "Due later",
				},
				{
					Task:        createTask("Task 5", "", "Due today"),
					SectionGID:  "section_Recently assigned",
					SectionName: "Recently assigned",
				},
			},
		},
		{
			name: "Tasks in ignored sections should not move",
			tasks: []asana.Task{
				createTask("Task 1", "2023-04-10", "Doing Now"),   // Past date, but in ignored section
				createTask("Task 2", "2023-05-15", "Waiting For"), // Future date, but in ignored section
			},
			config: core.SectionConfig{
				Overdue:     "Overdue",
				DueToday:    "Due today",
				DueThisWeek: "Due within the next 7 days",
				DueLater:    "Due later",
				NoDate:      "Recently assigned",
			},
			sectionNameToGID: map[string]string{
				"Overdue":                    "section_Overdue",
				"Due today":                  "section_Due today",
				"Due within the next 7 days": "section_Due within the next 7 days",
				"Due later":                  "section_Due later",
				"Recently assigned":          "section_Recently assigned",
			},
			ignoredSections: map[string]bool{
				"Doing Now":   true,
				"Waiting For": true,
			},
			expectedMoves: []core.TaskMove{}, // No moves expected
		},
		{
			name: "Tasks should not move to ignored target sections",
			tasks: []asana.Task{
				createTask("Task 1", "2023-04-15", "Custom Section"),  // Today's date, but Due today is ignored
				createTask("Task 2", "2023-04-10", "Another Section"), // Past date, but Overdue is ignored
			},
			config: core.SectionConfig{
				Overdue:     "Overdue",
				DueToday:    "Due today",
				DueThisWeek: "Due within the next 7 days",
				DueLater:    "Due later",
				NoDate:      "Recently assigned",
			},
			sectionNameToGID: map[string]string{
				"Overdue":                    "section_Overdue",
				"Due today":                  "section_Due today",
				"Due within the next 7 days": "section_Due within the next 7 days",
				"Due later":                  "section_Due later",
				"Recently assigned":          "section_Recently assigned",
			},
			ignoredSections: map[string]bool{
				"Overdue":   true,
				"Due today": true,
			},
			expectedMoves: []core.TaskMove{}, // No moves expected
		},
		{
			name: "Tasks should not move to sections that don't exist",
			tasks: []asana.Task{
				createTask("Task 1", "2023-04-15", "Custom Section"),  // Today's date
				createTask("Task 2", "2023-04-10", "Another Section"), // Past date
			},
			config: core.SectionConfig{
				Overdue:     "Overdue",
				DueToday:    "Due today",
				DueThisWeek: "Due within the next 7 days",
				DueLater:    "Due later",
				NoDate:      "Recently assigned",
			},
			sectionNameToGID: map[string]string{
				// Missing the Overdue and Due today sections
				"Due within the next 7 days": "section_Due within the next 7 days",
				"Due later":                  "section_Due later",
				"Recently assigned":          "section_Recently assigned",
			},
			ignoredSections: map[string]bool{},
			expectedMoves:   []core.TaskMove{}, // No moves expected since target sections don't exist
		},
		{
			name: "Edge cases: tasks at boundaries of date ranges",
			tasks: []asana.Task{
				// Edge of today and tomorrow
				createTask("Task 1", "2023-04-15", "Wrong Section"), // Last minute of today
				// Edge of week boundary
				createTask("Task 2", "2023-04-22", "Wrong Section"), // 7 days from reference (should be this week)
				createTask("Task 3", "2023-04-23", "Wrong Section"), // 8 days from reference (should be due later)
			},
			config: core.SectionConfig{
				Overdue:     "Overdue",
				DueToday:    "Due today",
				DueThisWeek: "Due within the next 7 days",
				DueLater:    "Due later",
				NoDate:      "Recently assigned",
			},
			sectionNameToGID: map[string]string{
				"Overdue":                    "section_Overdue",
				"Due today":                  "section_Due today",
				"Due within the next 7 days": "section_Due within the next 7 days",
				"Due later":                  "section_Due later",
				"Recently assigned":          "section_Recently assigned",
			},
			ignoredSections: map[string]bool{},
			expectedMoves: []core.TaskMove{
				{
					Task:        createTask("Task 1", "2023-04-15", "Wrong Section"),
					SectionGID:  "section_Due today",
					SectionName: "Due today",
				},
				{
					Task:        createTask("Task 2", "2023-04-22", "Wrong Section"),
					SectionGID:  "section_Due within the next 7 days",
					SectionName: "Due within the next 7 days",
				},
				{
					Task:        createTask("Task 3", "2023-04-23", "Wrong Section"),
					SectionGID:  "section_Due later",
					SectionName: "Due later",
				},
			},
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			moves := core.CalculateTaskMoves(tc.tasks, tc.config, tc.sectionNameToGID, tc.ignoredSections, referenceTime)

			// For test cases where we expect no moves, just check the length
			if len(tc.expectedMoves) == 0 && len(moves) == 0 {
				// Both are empty, this is correct
				return
			}

			// For test cases with expected moves, compare the details
			if len(moves) != len(tc.expectedMoves) {
				t.Errorf("Expected %d moves, got %d", len(tc.expectedMoves), len(moves))
				return
			}

			// Convert to a comparable format and check for equality
			// This approach avoids issues with direct struct comparisons
			for i, move := range moves {
				expected := tc.expectedMoves[i]

				if move.Task.GID != expected.Task.GID {
					t.Errorf("Move %d: Expected Task.GID=%s, got %s", i, expected.Task.GID, move.Task.GID)
				}

				if move.Task.Name != expected.Task.Name {
					t.Errorf("Move %d: Expected Task.Name=%s, got %s", i, expected.Task.Name, move.Task.Name)
				}

				if move.SectionGID != expected.SectionGID {
					t.Errorf("Move %d: Expected SectionGID=%s, got %s", i, expected.SectionGID, move.SectionGID)
				}

				if move.SectionName != expected.SectionName {
					t.Errorf("Move %d: Expected SectionName=%s, got %s", i, expected.SectionName, move.SectionName)
				}
			}
		})
	}
}

// Test that the TaskMove struct is correctly initialized
func TestTaskMoveStruct(t *testing.T) {
	task := asana.Task{
		GID:  "task_123",
		Name: "Test Task",
	}

	move := core.TaskMove{
		Task:        task,
		SectionGID:  "section_456",
		SectionName: "New Section",
	}

	if move.Task.GID != "task_123" {
		t.Errorf("Expected Task.GID to be 'task_123', got '%s'", move.Task.GID)
	}

	if move.SectionGID != "section_456" {
		t.Errorf("Expected SectionGID to be 'section_456', got '%s'", move.SectionGID)
	}

	if move.SectionName != "New Section" {
		t.Errorf("Expected SectionName to be 'New Section', got '%s'", move.SectionName)
	}
}

// Test handling of date calculations
func TestDateCalculations(t *testing.T) {
	// Setup
	referenceTime := time.Date(2023, 4, 15, 12, 0, 0, 0, time.UTC) // Saturday
	config := core.DefaultSectionConfig()
	sectionNameToGID := map[string]string{
		"Overdue":                    "section_Overdue",
		"Due today":                  "section_Due today",
		"Due within the next 7 days": "section_Due within the next 7 days",
		"Due later":                  "section_Due later",
		"Recently assigned":          "section_Recently assigned",
	}
	ignoredSections := map[string]bool{}

	testCases := []struct {
		dueDate          string
		currentSection   string
		expectedCategory asana.TaskCategory
		expectedSection  string
		shouldMove       bool
	}{
		// Overdue dates
		{"2023-04-14", "Wrong Section", asana.Overdue, "Overdue", true},
		{"2023-01-01", "Wrong Section", asana.Overdue, "Overdue", true},

		// Today
		{"2023-04-15", "Wrong Section", asana.DueToday, "Due today", true},

		// Within next 7 days
		{"2023-04-16", "Wrong Section", asana.DueThisWeek, "Due within the next 7 days", true}, // Tomorrow
		{"2023-04-22", "Wrong Section", asana.DueThisWeek, "Due within the next 7 days", true}, // 7 days from now

		// Due later
		{"2023-04-23", "Wrong Section", asana.DueLater, "Due later", true}, // 8 days from now
		{"2023-05-01", "Wrong Section", asana.DueLater, "Due later", true},
		{"2024-01-01", "Wrong Section", asana.DueLater, "Due later", true},

		// No date
		{"", "Wrong Section", asana.NoDate, "Recently assigned", true},

		// Already in correct section
		{"2023-04-14", "Overdue", asana.Overdue, "Overdue", false},
		{"2023-04-15", "Due today", asana.DueToday, "Due today", false},
		{"2023-04-20", "Due within the next 7 days", asana.DueThisWeek, "Due within the next 7 days", false},
		{"2023-05-01", "Due later", asana.DueLater, "Due later", false},
		{"", "Recently assigned", asana.NoDate, "Recently assigned", false},
	}

	for _, tc := range testCases {
		t.Run(tc.dueDate+"-"+tc.currentSection, func(t *testing.T) {
			// Create test task
			var task asana.Task
			if tc.dueDate == "" {
				task = asana.Task{
					GID:  "task_test",
					Name: "Test Task",
					AssigneeSection: asana.AssigneeSection{
						GID:  "section_" + tc.currentSection,
						Name: tc.currentSection,
					},
				}
			} else {
				dueDate, _ := time.Parse("2006-01-02", tc.dueDate)
				task = asana.Task{
					GID:   "task_test",
					Name:  "Test Task",
					DueOn: asana.Date(dueDate),
					AssigneeSection: asana.AssigneeSection{
						GID:  "section_" + tc.currentSection,
						Name: tc.currentSection,
					},
				}
			}

			// Calculate the actual category
			actualCategory := task.GetTaskCategory(referenceTime)
			if actualCategory != tc.expectedCategory {
				t.Errorf("Expected category %v for date %s, got %v",
					tc.expectedCategory, tc.dueDate, actualCategory)
			}

			// Calculate moves
			moves := core.CalculateTaskMoves([]asana.Task{task}, config, sectionNameToGID, ignoredSections, referenceTime)

			// Check if the task should move
			if tc.shouldMove && len(moves) == 0 {
				t.Errorf("Expected task to move for date %s from section %s to %s, but no move was calculated",
					tc.dueDate, tc.currentSection, tc.expectedSection)
			} else if !tc.shouldMove && len(moves) > 0 {
				t.Errorf("Expected task NOT to move for date %s from section %s, but a move was calculated",
					tc.dueDate, tc.currentSection)
			}

			// If it should move, verify the target section
			if tc.shouldMove && len(moves) > 0 {
				if moves[0].SectionName != tc.expectedSection {
					t.Errorf("Expected move to section %s for date %s, got %s",
						tc.expectedSection, tc.dueDate, moves[0].SectionName)
				}
			}
		})
	}
}
