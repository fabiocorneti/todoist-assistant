package todoist

//go:generate mockery --name=Transport --inpackage --structname=MockTransport
type Transport interface {
	getProjects() ([]Project, error)
	getAllTasks() ([]Task, error)
	getTasksForProject(projectID string) ([]Task, error)
	getTaskLabels(taskID string) ([]string, error)
	setTaskPriority(taskID string, priority int) error
	completeTask(taskID string) error
	createTask(content, projectID string) (*Task, error)
	updateTaskLabels(taskID string, labels []string) error
}
