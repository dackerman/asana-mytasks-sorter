package asana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// API constants
const (
	BaseURL        = "https://app.asana.com/api/1.0"
	DefaultTimeout = 10 * time.Second
	
	// Query parameter names
	QueryCompletedSince = "completed_since"
	QueryOptFields      = "opt_fields"
	QueryWorkspace      = "workspace"
	
	// Standard field sets
	TaskFields = "name,completed,due_on,due_at,assignee_section,assignee_section.name"
)

// Client handles API requests to the Asana API
type Client struct {
	Client  *http.Client
	Token   string
	BaseURL string
}

// NewClient creates a new Asana API client
func NewClient(token string) *Client {
	return &Client{
		Client:  &http.Client{Timeout: DefaultTimeout},
		Token:   token,
		BaseURL: BaseURL,
	}
}

// Request represents an API request
type Request struct {
	Method      string
	Path        string
	QueryParams map[string]string
	Body        interface{}
	Context     context.Context
}

// Response structs
type DataContainer struct {
	Data json.RawMessage `json:"data"`
}

// Model types
type User struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type Workspace struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type Section struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type AssigneeSection struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

type UserTaskList struct {
	GID       string    `json:"gid"`
	Name      string    `json:"name"`
	Owner     User      `json:"owner"`
	Workspace Workspace `json:"workspace"`
}

type Task struct {
	GID             string          `json:"gid"`
	Name            string          `json:"name"`
	Completed       bool            `json:"completed"`
	DueOn           Date            `json:"due_on,omitempty"`
	DueAt           time.Time       `json:"due_at,omitempty"`
	AssigneeSection AssigneeSection `json:"assignee_section,omitempty"`
}

// Date is a custom type to handle ISO 8601 date strings
type Date time.Time

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *Date) UnmarshalJSON(data []byte) error {
	// Handle null values
	if string(data) == "null" {
		*d = Date(time.Time{})
		return nil
	}

	// Strip quotes
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	// Empty string
	if s == "" {
		*d = Date(time.Time{})
		return nil
	}

	// Parse the ISO 8601 date
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}

	*d = Date(t)
	return nil
}

// Time returns the time.Time representation
func (d Date) Time() time.Time {
	return time.Time(d)
}

// Format formats the date using the given layout
func (d Date) Format(layout string) string {
	return time.Time(d).Format(layout)
}

// IsZero reports whether the date is zero
func (d Date) IsZero() bool {
	return time.Time(d).IsZero()
}

// TaskCategory represents a category for tasks based on due date
type TaskCategory int

const (
	Overdue TaskCategory = iota
	DueToday
	DueThisWeek
	DueLater
	NoDate
)

// GetTaskCategory determines the category of a task based on its due date
func (t *Task) GetTaskCategory(now time.Time) TaskCategory {
	if t.DueOn.IsZero() {
		return NoDate
	}

	dueDate := t.DueOn.Time()

	// Normalize time to start of day for comparison
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDateNormalized := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, now.Location())

	// Tasks due today should always be in the DueToday category
	if dueDateNormalized.Equal(nowDate) {
		return DueToday
	}

	// Compare dates
	if dueDateNormalized.Before(nowDate) {
		return Overdue
	} 
	
	// For future dates, calculate days difference
	days := int(dueDateNormalized.Sub(nowDate).Hours() / 24)
	if days <= 7 {
		return DueThisWeek
	} 
	
	return DueLater
}

// This section has been moved to internal/core/config.go

// executeRequest is a generic helper method for making HTTP requests to the Asana API
func (c *Client) executeRequest(req Request) ([]byte, error) {
	// Use background context if none provided
	ctx := req.Context
	if ctx == nil {
		ctx = context.Background()
	}
	
	// Build URL with query parameters
	reqURL, err := url.Parse(c.BaseURL + req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	
	// Add query parameters if any
	if len(req.QueryParams) > 0 {
		q := reqURL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, value)
		}
		reqURL.RawQuery = q.Encode()
	}
	
	// Create request body if any
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error creating request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}
	
	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, reqURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Add common headers
	httpReq.Header.Add("Authorization", "Bearer "+c.Token)
	httpReq.Header.Add("Accept", "application/json")
	
	// Add content-type for requests with bodies
	if req.Body != nil {
		httpReq.Header.Add("Content-Type", "application/json")
	}
	
	// Execute request
	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return bodyBytes, nil
}

// unmarshalResponse is a helper function to unmarshal the API response
func unmarshalResponse(data []byte, target interface{}) error {
	var container DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return fmt.Errorf("failed to unmarshal API response container: %w", err)
	}
	
	if err := json.Unmarshal(container.Data, target); err != nil {
		return fmt.Errorf("failed to unmarshal API response data: %w", err)
	}
	
	return nil
}

// GetCurrentUser retrieves the current user's information
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	data, err := c.executeRequest(Request{
		Method:  http.MethodGet,
		Path:    "/users/me",
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	
	var user User
	if err := unmarshalResponse(data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}
	
	return &user, nil
}

// GetWorkspaces retrieves all workspaces the user has access to
func (c *Client) GetWorkspaces(ctx context.Context) ([]Workspace, error) {
	data, err := c.executeRequest(Request{
		Method:  http.MethodGet,
		Path:    "/workspaces",
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}
	
	var workspaces []Workspace
	if err := unmarshalResponse(data, &workspaces); err != nil {
		return nil, fmt.Errorf("failed to parse workspaces: %w", err)
	}
	
	return workspaces, nil
}

// GetUserTaskList retrieves a user's task list for a specific workspace
func (c *Client) GetUserTaskList(ctx context.Context, userGID, workspaceGID string) (*UserTaskList, error) {
	data, err := c.executeRequest(Request{
		Method:  http.MethodGet,
		Path:    fmt.Sprintf("/users/%s/user_task_list", userGID),
		QueryParams: map[string]string{
			QueryWorkspace: workspaceGID,
		},
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user task list: %w", err)
	}
	
	var userTaskList UserTaskList
	if err := unmarshalResponse(data, &userTaskList); err != nil {
		return nil, fmt.Errorf("failed to parse user task list: %w", err)
	}
	
	return &userTaskList, nil
}

// GetSectionsForProject retrieves all sections in a project
func (c *Client) GetSectionsForProject(ctx context.Context, projectGID string) ([]Section, error) {
	data, err := c.executeRequest(Request{
		Method:  http.MethodGet,
		Path:    fmt.Sprintf("/projects/%s/sections", projectGID),
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get sections for project: %w", err)
	}
	
	var sections []Section
	if err := unmarshalResponse(data, &sections); err != nil {
		return nil, fmt.Errorf("failed to parse sections: %w", err)
	}
	
	return sections, nil
}

// GetTasksFromUserTaskList retrieves all incomplete tasks in a user's task list
func (c *Client) GetTasksFromUserTaskList(ctx context.Context, userTaskListGID string) ([]Task, error) {
	data, err := c.executeRequest(Request{
		Method:  http.MethodGet,
		Path:    fmt.Sprintf("/user_task_lists/%s/tasks", userTaskListGID),
		QueryParams: map[string]string{
			QueryCompletedSince: "now",
			QueryOptFields:      TaskFields,
		},
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks from user task list: %w", err)
	}
	
	var tasks []Task
	if err := unmarshalResponse(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}
	
	return tasks, nil
}

// GetTasksInSection retrieves all incomplete tasks in a section
func (c *Client) GetTasksInSection(ctx context.Context, sectionGID string) ([]Task, error) {
	data, err := c.executeRequest(Request{
		Method:  http.MethodGet,
		Path:    fmt.Sprintf("/sections/%s/tasks", sectionGID),
		QueryParams: map[string]string{
			QueryCompletedSince: "now",
			QueryOptFields:      TaskFields,
		},
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks in section: %w", err)
	}
	
	var tasks []Task
	if err := unmarshalResponse(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}
	
	return tasks, nil
}

// CreateSection creates a new section in a project
func (c *Client) CreateSection(ctx context.Context, projectGID, name string) (*Section, error) {
	data, err := c.executeRequest(Request{
		Method:  http.MethodPost,
		Path:    fmt.Sprintf("/projects/%s/sections", projectGID),
		Body:    map[string]interface{}{
			"data": map[string]string{
				"name": name,
			},
		},
		Context: ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}
	
	var section Section
	if err := unmarshalResponse(data, &section); err != nil {
		return nil, fmt.Errorf("failed to parse created section: %w", err)
	}
	
	return &section, nil
}

// MoveTaskToSection moves a task to a section
func (c *Client) MoveTaskToSection(ctx context.Context, sectionGID, taskGID string) error {
	_, err := c.executeRequest(Request{
		Method:  http.MethodPost,
		Path:    fmt.Sprintf("/sections/%s/addTask", sectionGID),
		Body:    map[string]interface{}{
			"data": map[string]string{
				"task": taskGID,
			},
		},
		Context: ctx,
	})
	
	if err != nil {
		return fmt.Errorf("failed to move task to section: %w", err)
	}
	
	return nil
}