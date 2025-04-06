package asana

import (
	"bytes"
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
	GID       string    `json:"gid"`
	Name      string    `json:"name"`
	Owner     User      `json:"owner"`
	Workspace Workspace `json:"workspace"`
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
	GID       string    `json:"gid"`
	Name      string    `json:"name"`
	Completed bool      `json:"completed"`
	DueOn     Date      `json:"due_on,omitempty"`
	DueAt     time.Time `json:"due_at,omitempty"`
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
	} else {
		// For future dates, calculate days difference
		days := int(dueDateNormalized.Sub(nowDate).Hours() / 24)
		if days <= 7 {
			return DueThisWeek
		} else {
			return DueLater
		}
	}
}

// SectionConfig defines the mapping of task categories to section names
type SectionConfig struct {
	Overdue         string   `json:"overdue"`
	DueToday        string   `json:"due_today"`
	DueThisWeek     string   `json:"due_this_week"`
	DueLater        string   `json:"due_later"`
	NoDate          string   `json:"no_date"`
	IgnoredSections []string `json:"ignored_sections,omitempty"`
}

// DefaultSectionConfig returns the default section configuration
func DefaultSectionConfig() SectionConfig {
	return SectionConfig{
		Overdue:         "Overdue",
		DueToday:        "Due today",
		DueThisWeek:     "Due within the next 7 days",
		DueLater:        "Due later",
		NoDate:          "Recently assigned",
		IgnoredSections: []string{},
	}
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
		"opt_fields":      "name,completed,due_on,due_at",
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
		"opt_fields":      "name,completed,due_on,due_at",
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

// CreateSection creates a new section in a project
func (c *Client) CreateSection(projectGID, name string) (*Section, error) {
	path := fmt.Sprintf("/projects/%s/sections", projectGID)

	// Create request body
	requestBody := map[string]interface{}{
		"data": map[string]string{
			"name": name,
		},
	}

	// Convert to JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request body: %v", err)
	}

	// Make the request
	data, err := c.makePostRequest(path, bodyBytes)
	if err != nil {
		return nil, err
	}

	var container DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	var section Section
	if err := json.Unmarshal(container.Data, &section); err != nil {
		return nil, err
	}

	return &section, nil
}

// MoveTaskToSection moves a task to a section
func (c *Client) MoveTaskToSection(sectionGID, taskGID string) error {
	path := fmt.Sprintf("/sections/%s/addTask", sectionGID)

	// Create request body
	requestBody := map[string]interface{}{
		"data": map[string]string{
			"task": taskGID,
		},
	}

	// Convert to JSON
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error creating request body: %v", err)
	}

	// Make the request
	_, err = c.makePostRequest(path, bodyBytes)
	return err
}

// makePostRequest is a helper method to make HTTP POST requests to the Asana API
func (c *Client) makePostRequest(path string, body []byte) ([]byte, error) {
	// Build URL
	reqURL := c.BaseURL + path

	// Create request
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}
