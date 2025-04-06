package asana

import "context"

// API defines the interface for interacting with the Asana API
type API interface {
	// User methods
	GetCurrentUser(ctx context.Context) (*User, error)
	GetWorkspaces(ctx context.Context) ([]Workspace, error)
	GetUserTaskList(ctx context.Context, userGID, workspaceGID string) (*UserTaskList, error)
	
	// Section methods
	GetSectionsForProject(ctx context.Context, projectGID string) ([]Section, error)
	CreateSection(ctx context.Context, projectGID, name string) (*Section, error)
	
	// Task methods
	GetTasksFromUserTaskList(ctx context.Context, userTaskListGID string) ([]Task, error)
	GetTasksInSection(ctx context.Context, sectionGID string) ([]Task, error)
	MoveTaskToSection(ctx context.Context, sectionGID, taskGID string) error
}

// Ensure Client implements the API interface
var _ API = (*Client)(nil)