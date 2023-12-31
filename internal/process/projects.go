package process

import (
	"strings"

	"github.com/fabiocorneti/todoist-assistant/internal/config"
	"github.com/fabiocorneti/todoist-assistant/internal/todoist"
	"github.com/fabiocorneti/todoist-assistant/internal/utils"
	"github.com/sirupsen/logrus"
)

type ProjectsProcess struct {
	config        config.Config
	logger        *logrus.Logger
	todoistClient *todoist.Client
	projects      []todoist.Project
}

func NewProjectsProcess(cfg config.Config, logger *logrus.Logger,
	todoistClient *todoist.Client, projects []todoist.Project) *ProjectsProcess {
	process := ProjectsProcess{
		config:        cfg,
		logger:        logger,
		todoistClient: todoistClient,
		projects:      projects,
	}
	return &process
}

func (process ProjectsProcess) ProcessProjects() {
	var err error
	var parentProjectID string
	parentProjectID, err = process.getParentProjectID()
	if err != nil {
		process.logger.Fatalf("Error locating parent project: %v", err)
	}

	process.logger.Info("Processing projects")
	for _, project := range process.projects {
		projectCopy := project
		if parentProjectID != "" && project.ParentID != parentProjectID {
			continue
		}
		process.logger.Debugf("Getting tasks for project %s (%s)", project.ID, project.Name)
		var tasks []todoist.Task
		tasks, err = process.todoistClient.GetTasksForProject(project.ID)
		if err != nil {
			process.logger.Fatalf("Error fetching Todoist tasks for project: %v", err)
			break
		}
		setNextAction := true
		for _, task := range tasks {
			taskCopy := task
			setNextAction = process.processTask(&projectCopy, &taskCopy, setNextAction)
		}
		process.logger.Infof("Completed processing of project %s (%s)", project.ID, project.Name)
	}
}

func (process ProjectsProcess) processTask(project *todoist.Project, task *todoist.Task, setNextAction bool) bool {
	label := process.config.Todoist.ProjectsLabelPrefix + "/" + project.Name
	if process.config.Todoist.AssignProjectLabel {
		process.logger.Debugf("Processing project task %s", task.Content)
		if !utils.Contains(task.Labels, label) {
			err := process.todoistClient.AddLabelsToTask(task.ID, []string{label})
			if err != nil {
				process.logger.Fatalf("Error adding project label to task %s", task.Content)
				return false
			}
		}
	}
	if !process.config.Todoist.AssignNextActionLabel {
		return false
	}
	if !setNextAction && utils.Contains(task.Labels, process.config.Todoist.NextActionLabel) {
		err := process.todoistClient.RemoveLabelsFromTask(task.ID, []string{process.config.Todoist.NextActionLabel})
		if err != nil {
			process.logger.Fatalf("Error removing next action label from task %s: %v", task.Content, err)
			return false
		}
		return false
	}

	// NOTE: do not set next action label on uncompletable tasks
	if setNextAction && !strings.HasPrefix(task.Content, "* ") {
		if !utils.Contains(task.Labels, process.config.Todoist.NextActionLabel) {
			err := process.todoistClient.AddLabelsToTask(task.ID, []string{process.config.Todoist.NextActionLabel})
			if err != nil {
				process.logger.Fatalf("Error adding next action label to task %s", task.Content)
				return false
			}
		}
		return false
	}
	return setNextAction
}

func (process ProjectsProcess) getParentProjectID() (string, error) {
	if process.config.Todoist.ParentProjectName != "" {
		process.logger.Debug("Getting parent project")
		parentProjectID, err := process.todoistClient.FindProjectID(process.projects,
			process.config.Todoist.ParentProjectName)
		if err != nil {
			return "", err
		}
		return parentProjectID, nil
	}
	return "", nil
}
