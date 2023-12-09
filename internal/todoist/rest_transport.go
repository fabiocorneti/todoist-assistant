package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	apiURL = "https://api.todoist.com/rest/v2/"
)

type RESTTodoistTransport struct {
	httpClient *RateLimitedClient
	token      string
	testMode   bool
}

func NewRESTTodoistTransport(token string, testMode bool) TodoistTransport {
	return &RESTTodoistTransport{
		httpClient: NewRateLimitedClient(),
		token:      token,
		testMode:   testMode,
	}
}

func (t *RESTTodoistTransport) getProjects() ([]Project, error) {
	req, err := t.newRequest("GET", apiURL+"projects", nil)
	if err != nil {
		return nil, err
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (t *RESTTodoistTransport) getTasksForProject(projectID string) ([]Task, error) {
	req, err := t.newRequest("GET", apiURL+"tasks?project_id="+projectID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (t *RESTTodoistTransport) getAllTasks() ([]Task, error) {
	req, err := t.newRequest("GET", apiURL+"tasks", nil)
	if err != nil {
		return nil, err
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (t *RESTTodoistTransport) getTaskLabels(taskID string) ([]string, error) {
	req, err := t.newRequest("GET", apiURL+"tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, err
	}

	return task.Labels, nil
}

func (t *RESTTodoistTransport) createTask(content, projectID string) (*Task, error) {
	task := Task{
		Content: content,
	}

	if projectID != "" {
		task.ProjectID = projectID
	}

	jsonTask, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	req, err := t.newRequest("POST", apiURL+"tasks", bytes.NewBuffer(jsonTask))
	if err != nil {
		return nil, err
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var createdTask Task
	if err := json.NewDecoder(resp.Body).Decode(&createdTask); err != nil {
		return nil, err
	}

	return &createdTask, nil
}

func (t *RESTTodoistTransport) updateTaskLabels(taskID string, labels []string) error {
	jsonData, err := json.Marshal(map[string][]string{"labels": labels})
	if err != nil {
		return err
	}

	req, err := t.newRequest("POST", apiURL+"tasks/"+taskID, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	_, err = t.httpClient.Do(req)
	return err
}

func (t *RESTTodoistTransport) setTaskPriority(taskID string, priority int) error {
	jsonData, err := json.Marshal(map[string]int{"priority": priority})
	if err != nil {
		return err
	}

	req, err := t.newRequest("POST", apiURL+"tasks/"+taskID, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	_, err = t.httpClient.Do(req)
	return err
}

func (t *RESTTodoistTransport) completeTask(taskID string) error {
	req, err := t.newRequest("POST", apiURL+"tasks/"+taskID+"/close", nil)
	if err != nil {
		return err
	}

	_, err = t.httpClient.Do(req)
	return err
}

func (t *RESTTodoistTransport) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	if t.testMode {
		log.Fatal("Cannot send requests in test mode")
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
