package process

import (
	"time"

	"github.com/fabiocorneti/todoist-assistant/internal/config"
	"github.com/fabiocorneti/todoist-assistant/internal/todoist"
	"github.com/sirupsen/logrus"
)

func RunProcess(cfg config.Config, logger *logrus.Logger) {
	start := time.Now()
	todoistClient := todoist.NewTodoistClient(cfg.Todoist.Token, cfg.IsTest())

	logger.Debug("Getting projects")
	projects, err := todoistClient.GetProjects()
	if err != nil {
		logger.Fatalf("Error fetching Todoist projects: %v", err)
	}

	if len(cfg.Jira) > 0 {
		jiraProcess := NewJiraProcess(cfg, logger, todoistClient, projects)
		jiraProcess.ProcessJiraInstances()
	}

	projectsProcess := NewProjectsProcess(cfg, logger, todoistClient, projects)
	projectsProcess.ProcessProjects()
	logger.Infof("Completed update in %f seconds", time.Since(start).Seconds())
}
