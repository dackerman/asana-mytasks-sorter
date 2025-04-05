package asana

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BaseURL is the base URL for the Asana API
const BaseURL = "https://app.asana.com/api/1.0"

// Client handles API requests to the Asana API
type Client struct {
	Client  *http.Client
	Token   string
	BaseURL string
}

// Response structs
type DataContainer struct {
	Data json.RawMessage `json:"data"`
}

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

type UserTaskList struct {
	GID        string `json:"gid"`
	Name       string `json:"name"`
	Owner      User   `json:"owner"`
	Workspace  Workspace `json:"workspace"`
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

type Task struct {
	GID        string    `json:"gid"`
	Name       string    `json:"name"`
	Notes      string    `json:"notes,omitempty"`
	Completed  bool      `json:"completed"`
	DueOn      Date      `json:"due_on,omitempty"`
	DueAt      time.Time `json:"due_at,omitempty"`
}

// NewClient creates a new Asana API client
func NewClient(token string) *Client {
	return &Client{
		Client:  &http.Client{Timeout: 10 * time.Second},
		Token:   token,
		BaseURL: BaseURL,
	}
}

// makeRequest is a helper method to make HTTP requests to the Asana API
func (c *Client) makeRequest(method, path string, queryParams map[string]string) ([]byte, error) {
	// Build URL with query parameters
	reqURL := c.BaseURL + path
	if queryParams != nil && len(queryParams) > 0 {
		reqURL += "?"
		for key, value := range queryParams {
			reqURL += key + "=" + value + "&"
		}
		// Remove trailing "&"
		reqURL = reqURL[:len(reqURL)-1]
	}

	// Create request
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Accept", "application/json")

	// Execute request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

// GetCurrentUser retrieves the current user's information
func (c *Client) GetCurrentUser() (*User, error) {
	data, err := c.makeRequest("GET", "/users/me", nil)
	if err != nil {
		return nil, err
	}

	var container DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(container.Data, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserTaskList retrieves a user's task list for a specific workspace
func (c *Client) GetUserTaskList(userGID, workspaceGID string) (*UserTaskList, error) {
	path := fmt.Sprintf("/users/%s/user_task_list", userGID)
	queryParams := map[string]string{
		"workspace": workspaceGID,
	}
	
	data, err := c.makeRequest("GET", path, queryParams)
	if err != nil {
		return nil, err
	}

	var container DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	var userTaskList UserTaskList
	if err := json.Unmarshal(container.Data, &userTaskList); err != nil {
		return nil, err
	}

	return &userTaskList, nil
}

// GetWorkspaces retrieves all workspaces the user has access to
func (c *Client) GetWorkspaces() ([]Workspace, error) {
	data, err := c.makeRequest("GET", "/workspaces", nil)
	if err != nil {
		return nil, err
	}

	var container DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	var workspaces []Workspace
	if err := json.Unmarshal(container.Data, &workspaces); err != nil {
		return nil, err
	}

	return workspaces, nil
}

// GetSectionsForProject retrieves all sections in a project
func (c *Client) GetSectionsForProject(projectGID string) ([]Section, error) {
	path := fmt.Sprintf("/projects/%s/sections", projectGID)
	
	data, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var container DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	var sections []Section
	if err := json.Unmarshal(container.Data, &sections); err != nil {
		return nil, err
	}

	return sections, nil
}

// GetTasksInSection retrieves all incomplete tasks in a section
func (c *Client) GetTasksInSection(sectionGID string) ([]Task, error) {
	path := fmt.Sprintf("/sections/%s/tasks", sectionGID)
	
	queryParams := map[string]string{
		"completed_since": "now", // Only get incomplete tasks
		"opt_fields": "name,notes,completed,due_on,due_at",
	}

	data, err := c.makeRequest("GET", path, queryParams)
	if err != nil {
		return nil, err
	}

	var container DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(container.Data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetTasksFromUserTaskList retrieves all incomplete tasks in a user's task list
func (c *Client) GetTasksFromUserTaskList(userTaskListGID string) ([]Task, error) {
	path := fmt.Sprintf("/user_task_lists/%s/tasks", userTaskListGID)
	
	queryParams := map[string]string{
		"completed_since": "now",
		"opt_fields": "name,notes,completed,due_on,due_at",
	}

	data, err := c.makeRequest("GET", path, queryParams)
	if err != nil {
		return nil, err
	}

	var container DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(container.Data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}