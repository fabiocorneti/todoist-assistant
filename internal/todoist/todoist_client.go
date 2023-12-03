package todoist

import (
	"fmt"
)

type TodoistClient struct {
	transport TodoistTransport
}

func NewTodoistClient(token string, testMode bool) *TodoistClient {
	return &TodoistClient{
		transport: NewRESTTodoistTransport(token, testMode),
	}
}

func (tc *TodoistClient) GetProjects() ([]Project, error) {
	return tc.transport.getProjects()
}

func (tc *TodoistClient) FindProjectID(projects []Project, name string) (string, error) {
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

func (tc *TodoistClient) GetAllTasks() ([]Task, error) {
	return tc.transport.getAllTasks()
}

func (tc *TodoistClient) GetTasksForProject(projectID string) ([]Task, error) {
	return tc.transport.getTasksForProject(projectID)
}

func (tc *TodoistClient) CreateTask(content, projectID string) (*Task, error) {
	return tc.transport.createTask(content, projectID)
}

func (tc *TodoistClient) CompleteTask(taskID string) error {
	return tc.transport.completeTask(taskID)
}

func (tc *TodoistClient) ReplaceTaskLabels(taskID string, labels []string) error {
	return tc.transport.updateTaskLabels(taskID, labels)
}

func (tc *TodoistClient) AddLabelsToTask(taskID string, labels []string) error {
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

func (tc *TodoistClient) RemoveLabelsFromTask(taskID string, labels []string) error {
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
