package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fabiocorneti/todoist-assistant/internal"
	"github.com/fabiocorneti/todoist-assistant/internal/todoist"
	"github.com/sirupsen/logrus"
)

const (
	jiraMatches = 2
)

func main() {
	cfg := internal.GetConfiguration()
	logger := internal.GetLogger()

	ticker := time.NewTicker(time.Duration(cfg.UpdateInterval) * time.Minute)
	defer ticker.Stop()

	runProcess(cfg, logger)
	logger.Infof("Waiting %d minutes to perform the next update", cfg.UpdateInterval)
	for range ticker.C {
		runProcess(cfg, logger)
		logger.Infof("Waiting %d minutes to perform the next update", cfg.UpdateInterval)
	}
}

func runProcess(cfg internal.Config, logger *logrus.Logger) {
	var err error

	start := time.Now()
	processedTasks := make(map[string]todoist.Task)
	regexPattern := `\[([A-Z0-9-]+)\] .+\]\(.+\)$`
	re := regexp.MustCompile(regexPattern)

	todoistClient := todoist.NewTodoistClient(cfg.Todoist.Token, cfg.IsTest())

	logger.Debug("Getting projects")
	projects, err := todoistClient.GetProjects()
	if err != nil {
		logger.Fatalf("Error fetching Todoist projects: %v", err)
	}

	var tasks []todoist.Task
	var newTask *todoist.Task
	var jiraIssues []internal.JiraIssue
	if len(cfg.Jira) > 0 {
		logger.Info("Fetching Todoist tasks")
		tasks, err = todoistClient.GetAllTasks()
		if err != nil {
			logger.Errorf("Error fetching Todoist tasks: %v", err)
			return
		}

		logger.Debug("Finding tasks already linked to Jira issues")
		for _, task := range tasks {
			match := re.FindStringSubmatch(task.Content)
			if len(match) == jiraMatches {
				processedTasks[match[1]] = task
			}
		}
		logger.Debug("Issues already in Todoist:")
		for key := range processedTasks {
			logger.Debug(key)
		}

		for _, jiraConfig := range cfg.Jira {
			var targetProjectID string
			if jiraConfig.Project != "" {
				targetProjectID, err = todoistClient.FindProjectID(projects, jiraConfig.Project)
				if err != nil {
					logger.Fatalf("An error occurred when finding the target project for instance %s: %v", jiraConfig.Site, err)
				}
			}

			logger.Infof("Fetching issues from Jira instance %s", jiraConfig.Site)
			jiraIssues, err = internal.FetchJiraIssues(jiraConfig)
			if err != nil {
				logger.Fatalf("Error fetching Jira issues: %v", err)
				continue
			}

			for _, issue := range jiraIssues {
				var task todoist.Task
				var labelsToAdd []string
				labelsToAdd = append(labelsToAdd, jiraConfig.Labels...)
				taskPriority := 1
				for priority, jiraPriorityNames := range jiraConfig.PriorityMap {
					for _, name := range jiraPriorityNames {
						if issue.Fields.Priority.Name == name {
							taskPriority, err = cfg.ToAPIPriority(priority)
							if err != nil {
								logger.Fatalf(err.Error())
							}
							break
						}
					}
				}
				if jiraConfig.SyncJiraLabels {
					for _, label := range issue.Fields.Labels {
						labelsToAdd = append(labelsToAdd, fmt.Sprintf("Jira/Label/%s", label))
					}
				}
				if jiraConfig.SyncJiraComponents {
					for _, component := range issue.Fields.Components {
						labelsToAdd = append(labelsToAdd, fmt.Sprintf("Jira/Component/%s", component.Name))
					}
				}
				if _, exists := processedTasks[issue.Key]; !exists {
					if internal.Contains(jiraConfig.CompletionStatuses, issue.Fields.Status.Name) {
						logger.Debugf("Skipping completed issue [%s] %s", issue.Key, issue.Fields.Summary)
						continue
					}
					taskContent := internal.FormatTodoistTaskContent(jiraConfig, issue)
					newTask, err = todoistClient.CreateTask(taskContent, targetProjectID)
					if err != nil {
						logger.Fatalf("Error creating Todoist task: %v", err)
						break
					}
					logger.Infof("Created Todoist task: %v", taskContent)
					task = *newTask
				} else {
					task = processedTasks[issue.Key]
					logger.Debugf("Todoist task already exists for Jira issue [%s]", issue.Key)
					if internal.Contains(jiraConfig.CompletionStatuses, issue.Fields.Status.Name) {
						logger.Infof("Completing task %s", task.Content)
						err = todoistClient.CompleteTask(task.ID)
						if err != nil {
							logger.Fatalf("Error completing task %s: %v", task.Content, err)
							break
						}
						logger.Infof("Completed task %s", task.Content)
					}
				}

				logger.Debugf("Task %s priority: %d", task.Content, *task.Priority)
				if task.Priority != &taskPriority {
					logger.Debugf("Setting priority to %d for task %s", taskPriority, task.Content)
					err = todoistClient.SetTaskPriority(task.ID, taskPriority)
					if err != nil {
						logger.Fatalf("Error setting priority for task %s: %v", task.Content, err)
						break
					}
				}

				labelMap := make(map[string]bool)

				for _, label := range task.Labels {
					labelMap[label] = true
				}

				for _, label := range task.Labels {
					if strings.Index(label, "Jira/") == 0 {
						if !internal.Contains(labelsToAdd, label) {
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
					if internal.HaveSameElements(newLabels, task.Labels) {
						logger.Debugf("No need to sync Jira labels for task %s", task.Content)
					} else {
						err = todoistClient.ReplaceTaskLabels(task.ID, newLabels)
						if err != nil {
							logger.Fatalf("Error syncing Jira labels for task %s: %v", task.Content, err)
							break
						}
					}
				}
			}
			logger.Infof("Finished processing Jira instance %s", jiraConfig.Site)
		}
	}

	processProjects(*todoistClient, cfg, logger, projects)
	logger.Infof("Completed update in %f seconds", time.Since(start).Seconds())
}

func processProjects(todoist todoist.Client, cfg internal.Config, logger *logrus.Logger, projects []todoist.Project) {
	var err error
	var parentProjectID string
	if cfg.Todoist.ParentProjectName != "" {
		logger.Debug("Getting parent project")
		parentProjectID, err = todoist.FindProjectID(projects, cfg.Todoist.ParentProjectName)
		if err != nil {
			logger.Fatalf("Error locating parent project: %v", err)
		}
	}

	logger.Info("Processing projects")
	for _, project := range projects {
		if parentProjectID != "" && project.ParentID != parentProjectID {
			continue
		}
		logger.Debugf("Getting tasks for project %s (%s)", project.ID, project.Name)
		tasks, err := todoist.GetTasksForProject(project.ID)
		if err != nil {
			logger.Fatalf("Error fetching Todoist tasks for project: %v", err)
			break
		}
		label := cfg.Todoist.ProjectsLabelPrefix + "/" + project.Name
		setNextAction := true
		for _, task := range tasks {
			if cfg.Todoist.AssignProjectLabel {
				logger.Debugf("Processing project task %s", task.Content)
				if !internal.Contains(task.Labels, label) {
					err = todoist.AddLabelsToTask(task.ID, []string{label})
					if err != nil {
						logger.Fatalf("Error adding project label to task %s", task.Content)
						break
					}
				}
			}
			if cfg.Todoist.AssignNextActionLabel {
				// NOTE: do not set next action label on uncompletable tasks
				if setNextAction && !strings.HasPrefix(task.Content, "* ") {
					if !internal.Contains(task.Labels, cfg.Todoist.NextActionLabel) {
						err = todoist.AddLabelsToTask(task.ID, []string{cfg.Todoist.NextActionLabel})
						if err != nil {
							logger.Fatalf("Error adding next action label to task %s", task.Content)
							break
						}
					}
					setNextAction = false
				} else if internal.Contains(task.Labels, cfg.Todoist.NextActionLabel) {
					err = todoist.RemoveLabelsFromTask(task.ID, []string{cfg.Todoist.NextActionLabel})
					if err != nil {
						logger.Fatalf("Error removing next action label from task %s: %v", task.Content, err)
						break
					}
				}
			}
		}
		logger.Infof("Completed processing of project %s (%s)", project.ID, project.Name)
	}
}
