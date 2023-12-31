package process

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fabiocorneti/todoist-assistant/internal/config"
	"github.com/fabiocorneti/todoist-assistant/internal/jira"
	"github.com/fabiocorneti/todoist-assistant/internal/todoist"
	"github.com/fabiocorneti/todoist-assistant/internal/utils"
	"github.com/sirupsen/logrus"
)

const (
	jiraMatches = 2
)

type JiraProcess struct {
	config        config.Config
	logger        *logrus.Logger
	todoistClient *todoist.Client
	projects      []todoist.Project
}

func NewJiraProcess(cfg config.Config, logger *logrus.Logger,
	todoistClient *todoist.Client, projects []todoist.Project) *JiraProcess {
	process := JiraProcess{
		config:        cfg,
		logger:        logger,
		todoistClient: todoistClient,
		projects:      projects,
	}
	return &process
}

func (process JiraProcess) ProcessJiraInstances() {
	var err error

	processedTasks := make(map[string]todoist.Task)
	regexPattern := `\[([A-Z0-9-]+)\] .+\]\(.+\)$`
	re := regexp.MustCompile(regexPattern)

	var tasks []todoist.Task

	process.logger.Info("Fetching Todoist tasks")
	tasks, err = process.todoistClient.GetAllTasks()
	if err != nil {
		process.logger.Errorf("Error fetching Todoist tasks: %v", err)
		return
	}

	process.logger.Debug("Finding tasks already linked to Jira issues")
	for _, task := range tasks {
		match := re.FindStringSubmatch(task.Content)
		if len(match) == jiraMatches {
			processedTasks[match[1]] = task
		}
	}
	process.logger.Debug("Issues already in Todoist:")
	for key := range processedTasks {
		process.logger.Debug(key)
	}

	for _, jiraConfig := range process.config.Jira {
		process.processJiraInstance(jiraConfig, &processedTasks)
		process.logger.Infof("Finished processing Jira instance %s", jiraConfig.Site)
	}
}

func (process JiraProcess) processJiraInstance(jiraConfig config.JiraConfig, processedTasks *map[string]todoist.Task) {
	var err error
	var jiraIssues []jira.Issue

	var targetProjectID string
	if jiraConfig.Project != "" {
		targetProjectID, err = process.todoistClient.FindProjectID(process.projects, jiraConfig.Project)
		if err != nil {
			process.logger.Fatalf("An error occurred when finding the target project for instance %s: %v", jiraConfig.Site, err)
			return
		}
	}

	process.logger.Infof("Fetching issues from Jira instance %s", jiraConfig.Site)
	jiraIssues, err = jira.FetchJiraIssues(jiraConfig)
	if err != nil {
		process.logger.Fatalf("Error fetching Jira issues: %v", err)
		return
	}

	for _, issue := range jiraIssues {
		arg := issue
		process.processJiraIssue(jiraConfig, &arg, processedTasks, targetProjectID)
	}
}

func (process JiraProcess) processJiraIssue(jiraConfig config.JiraConfig, issue *jira.Issue,
	processedTasks *map[string]todoist.Task, targetProjectID string) {
	var err error
	task := process.getOrCreateTask(jiraConfig, issue, processedTasks, targetProjectID)

	if task == nil {
		return
	}

	taskPriority, err := process.getPriority(jiraConfig, issue)
	if err != nil {
		process.logger.Fatalf(err.Error())
		return
	}
	process.setTaskPriority(task, taskPriority)

	process.processLabels(jiraConfig, issue, task)
}

func (process JiraProcess) getOrCreateTask(jiraConfig config.JiraConfig, issue *jira.Issue,
	processedTasks *map[string]todoist.Task, targetProjectID string) *todoist.Task {
	if _, exists := (*processedTasks)[issue.Key]; !exists {
		if utils.Contains(jiraConfig.CompletionStatuses, issue.Fields.Status.Name) {
			process.logger.Debugf("Skipping completed issue [%s] %s", issue.Key, issue.Fields.Summary)
			return nil
		}
		taskContent := utils.FormatTodoistTaskContent(jiraConfig, *issue)
		task, err := process.todoistClient.CreateTask(taskContent, targetProjectID)
		if err != nil {
			process.logger.Fatalf("Error creating Todoist task: %v", err)
			return nil
		}
		process.logger.Infof("Created Todoist task: %v", taskContent)
		return task
	}

	task := (*processedTasks)[issue.Key]
	process.logger.Debugf("Todoist task already exists for Jira issue [%s]", issue.Key)
	if utils.Contains(jiraConfig.CompletionStatuses, issue.Fields.Status.Name) {
		process.logger.Infof("Completing task %s", task.Content)
		err := process.todoistClient.CompleteTask(task.ID)
		if err != nil {
			process.logger.Fatalf("Error completing task %s: %v", task.Content, err)
			return nil
		}
		process.logger.Infof("Completed task %s", task.Content)
	}
	return &task
}

func (process JiraProcess) setTaskPriority(task *todoist.Task, priority int) {
	process.logger.Debugf("Task %s priority: %d", task.Content, *task.Priority)
	if task.Priority != &priority {
		process.logger.Debugf("Setting priority to %d for task %s", priority, task.Content)
		err := process.todoistClient.SetTaskPriority(task.ID, priority)
		if err != nil {
			process.logger.Fatalf("Error setting priority for task %s: %v", task.Content, err)
			return
		}
	}
}

func (process JiraProcess) processLabels(cfg config.JiraConfig, issue *jira.Issue, task *todoist.Task) {
	labelsToAdd := process.collectLabelsToAdd(cfg, issue)

	labelMap := make(map[string]bool)
	for _, label := range task.Labels {
		labelMap[label] = true
	}
	for _, label := range task.Labels {
		if strings.Index(label, "Jira/") == 0 {
			if !utils.Contains(labelsToAdd, label) {
				delete(labelMap, label)
			}
		}
	}

	for _, label := range labelsToAdd {
		labelMap[label] = true
	}

	newLabels := make([]string, 0, len(labelMap))
	for label := range labelMap {
		newLabels = append(newLabels, label)
	}

	if len(newLabels) > 0 {
		if utils.HaveSameElements(newLabels, task.Labels) {
			process.logger.Debugf("No need to sync Jira labels for task %s", task.Content)
		} else {
			err := process.todoistClient.ReplaceTaskLabels(task.ID, newLabels)
			if err != nil {
				process.logger.Fatalf("Error syncing Jira labels for task %s: %v", task.Content, err)
				return
			}
		}
	}
}

func (process JiraProcess) collectLabelsToAdd(cfg config.JiraConfig, issue *jira.Issue) []string {
	var labelsToAdd []string
	labelsToAdd = append(labelsToAdd, cfg.Labels...)

	if cfg.SyncJiraLabels {
		for _, label := range issue.Fields.Labels {
			labelsToAdd = append(labelsToAdd, fmt.Sprintf("Jira/Label/%s", label))
		}
	}
	if cfg.SyncJiraComponents {
		for _, component := range issue.Fields.Components {
			labelsToAdd = append(labelsToAdd, fmt.Sprintf("Jira/Component/%s", component.Name))
		}
	}

	return labelsToAdd
}

func (process JiraProcess) getPriority(cfg config.JiraConfig, issue *jira.Issue) (int, error) {
	for priority, jiraPriorityNames := range cfg.PriorityMap {
		for _, name := range jiraPriorityNames {
			if issue.Fields.Priority.Name == name {
				taskPriority, err := process.config.ToAPIPriority(priority)
				if err != nil {
					return 0, err
				}
				return taskPriority, nil
			}
		}
	}
	return 1, nil
}
