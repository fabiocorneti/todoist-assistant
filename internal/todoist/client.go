package todoist

import (
	"fmt"
)

type Client struct {
	transport Transport
}

func NewTodoistClient(token string, testMode bool) *Client {
	return &Client{
		transport: NewRESTTodoistTransport(token, testMode),
	}
}

func (tc *Client) GetProjects() ([]Project, error) {
	return tc.transport.getProjects()
}

func (tc *Client) FindProjectID(projects []Project, name string) (string, error) {
	var id string
	for _, p := range projects {
		if p.Name == name {
			if id != "" {
				return "", fmt.Errorf("found more than one project for name %s", name)
			}
			id = p.ID
		}
	}
	if id == "" {
		return "", fmt.Errorf("project %s not found", name)
	}
	return id, nil
}

func (tc *Client) GetAllTasks() ([]Task, error) {
	return tc.transport.getAllTasks()
}

func (tc *Client) GetTasksForProject(projectID string) ([]Task, error) {
	return tc.transport.getTasksForProject(projectID)
}

func (tc *Client) CreateTask(content, projectID string) (*Task, error) {
	return tc.transport.createTask(content, projectID)
}

func (tc *Client) CompleteTask(taskID string) error {
	return tc.transport.completeTask(taskID)
}

func (tc *Client) ReplaceTaskLabels(taskID string, labels []string) error {
	return tc.transport.updateTaskLabels(taskID, labels)
}

func (tc *Client) SetTaskPriority(taskID string, priority int) error {
	return tc.transport.setTaskPriority(taskID, priority)
}

func (tc *Client) AddLabelsToTask(taskID string, labels []string) error {
	taskLabels, err := tc.transport.getTaskLabels(taskID)
	if err != nil {
		return err
	}

	for _, label := range labels {
		if !Contains(taskLabels, label) {
			taskLabels = append(taskLabels, label)
		}
	}
	return tc.transport.updateTaskLabels(taskID, taskLabels)
}

func (tc *Client) RemoveLabelsFromTask(taskID string, labels []string) error {
	currentLabels, err := tc.transport.getTaskLabels(taskID)
	if err != nil {
		return err
	}

	var newLabels []string
	for _, label := range currentLabels {
		if !Contains(labels, label) {
			newLabels = append(newLabels, label)
		}
	}

	return tc.transport.updateTaskLabels(taskID, newLabels)
}

func Contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
