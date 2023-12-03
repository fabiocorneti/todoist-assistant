package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fabiocorneti/todoist-assistant/internal"
	"github.com/fabiocorneti/todoist-assistant/internal/todoist"
)

func main() {
	cfg := internal.Configuration

	ticker := time.NewTicker(time.Duration(cfg.UpdateInterval) * time.Minute)
	defer ticker.Stop()

	runProcess()
	internal.Log.Infof("Waiting %d minutes to perform the next update", cfg.UpdateInterval)
	for range ticker.C {
		runProcess()
		internal.Log.Infof("Waiting %d minutes to perform the next update", cfg.UpdateInterval)
	}
}

func runProcess() {

	start := time.Now()
	processedTasks := make(map[string]todoist.Task)
	regexPattern := `\[([A-Z0-9-]+)\] .+\]\(.+\)$`
	re := regexp.MustCompile(regexPattern)

	todoistClient := todoist.NewTodoistClient(internal.Configuration.Todoist.Token, internal.Configuration.IsTest())

	internal.Log.Debug("Getting projects")
	projects, err := todoistClient.GetProjects()
	if err != nil {
		internal.Log.Fatalf("Error fetching Todoist projects: %v", err)
	}

	if len(internal.Configuration.Jira) > 0 {
		internal.Log.Info("Fetching Todoist tasks")
		tasks, err := todoistClient.GetAllTasks()
		if err != nil {
			internal.Log.Errorf("Error fetching Todoist tasks: %v", err)
			return
		}

		internal.Log.Debug("Finding tasks already linked to Jira issues")
		for _, task := range tasks {
			match := re.FindStringSubmatch(task.Content)
			if len(match) == 2 {
				processedTasks[match[1]] = task
			}
		}
		internal.Log.Debug("Issues already in Todoist:")
		for key := range processedTasks {
			internal.Log.Debug(key)
		}

		for _, jiraConfig := range internal.Configuration.Jira {
			var targetProjectID string
			if jiraConfig.Project != "" {
				targetProjectID, err = todoistClient.FindProjectID(projects, jiraConfig.Project)
				if err != nil {
					internal.Log.Fatalf("An error occurred when finding the target project for instance %s: %v", jiraConfig.Site, err)
				}
			}

			internal.Log.Infof("Fetching issues from Jira instance %s", jiraConfig.Site)
			jiraIssues, err := internal.FetchJiraIssues(jiraConfig)
			if err != nil {
				internal.Log.Fatalf("Error fetching Jira issues: %v", err)
				continue
			}

			for _, issue := range jiraIssues {
				var task todoist.Task
				var labelsToAdd []string
				labelsToAdd = append(labelsToAdd, jiraConfig.Labels...)
				if jiraConfig.SyncJiraLabels {
					for _, label := range issue.Fields.Labels {
						labelsToAdd = append(labelsToAdd, fmt.Sprintf("Jira/Label/%s", label))
					}
				}
				if jiraConfig.SyncJiraComponents {
					for _, component := range issue.Fields.Component {
						labelsToAdd = append(labelsToAdd, fmt.Sprintf("Jira/Component/%s", component.Name))
					}
				}
				if _, exists := processedTasks[issue.Key]; !exists {
					if internal.Contains(jiraConfig.CompletionStatuses, issue.Fields.Status.Name) {
						internal.Log.Debugf("Skipping completed issue [%s] %s", issue.Key, issue.Fields.Summary)
						continue
					}
					taskContent := internal.FormatTodoistTaskContent(jiraConfig, issue)
					newTask, err := todoistClient.CreateTask(taskContent, targetProjectID)
					if err != nil {
						internal.Log.Fatalf("Error creating Todoist task: %v", err)
						break
					} else {
						internal.Log.Infof("Created Todoist task: %v", taskContent)
					}
					task = *newTask
				} else {
					task = processedTasks[issue.Key]
					internal.Log.Debugf("Todoist task already exists for Jira issue [%s]", issue.Key)
					if internal.Contains(jiraConfig.CompletionStatuses, issue.Fields.Status.Name) {
						internal.Log.Infof("Completing task %s", task.Content)
						err = todoistClient.CompleteTask(task.ID)
						if err != nil {
							internal.Log.Fatalf("Error completing task %s: %v", task.Content, err)
							break
						}
						internal.Log.Infof("Completed task %s", task.Content)
						continue
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
						internal.Log.Debugf("No need to sync Jira labels for task %s", task.Content)
						continue
					}
					err = todoistClient.ReplaceTaskLabels(task.ID, newLabels)
					if err != nil {
						internal.Log.Fatalf("Error syncing Jira labels for task %s: %v", task.Content, err)
						break
					}
				}
			}
			internal.Log.Infof("Finished processing Jira instance %s", jiraConfig.Site)
		}
	}

	var parentProjectID string
	if internal.Configuration.Todoist.ParentProjectName != "" {
		internal.Log.Debug("Getting parent project")
		parentProjectID, err = todoistClient.FindProjectID(projects, internal.Configuration.Todoist.ParentProjectName)
		if err != nil {
			internal.Log.Fatalf("Error locating parent project: %v", err)
		}
	}

	internal.Log.Info("Processing projects")
	for _, project := range projects {
		if parentProjectID != "" && project.ParentID != parentProjectID {
			continue
		}
		internal.Log.Debugf("Getting tasks for project %s (%s)", project.ID, project.Name)
		tasks, err := todoistClient.GetTasksForProject(project.ID)
		if err != nil {
			internal.Log.Fatalf("Error fetching Todoist tasks for project: %v", err)
			break
		}
		label := internal.Configuration.Todoist.ProjectsLabelPrefix + "/" + project.Name
		setNextAction := true
		for _, task := range tasks {
			if internal.Configuration.Todoist.AssignProjectLabel {
				internal.Log.Debugf("Processing project task %s", task.Content)
				if !internal.Contains(task.Labels, label) {
					err = todoistClient.AddLabelsToTask(task.ID, []string{label})
					if err != nil {
						internal.Log.Fatalf("Error adding project label to task %s", task.Content)
						break
					}
				}
			}
			if internal.Configuration.Todoist.AssignNextActionLabel {
				// NOTE: do not set next action label on uncompletable tasks
				if setNextAction && !strings.HasPrefix(task.Content, "* ") {
					if !internal.Contains(task.Labels, internal.Configuration.Todoist.NextActionLabel) {
						err = todoistClient.AddLabelsToTask(task.ID, []string{internal.Configuration.Todoist.NextActionLabel})
						if err != nil {
							internal.Log.Fatalf("Error adding next action label to task %s", task.Content)
							break
						}
					}
					setNextAction = false
				} else {
					if internal.Contains(task.Labels, internal.Configuration.Todoist.NextActionLabel) {
						err = todoistClient.RemoveLabelsFromTask(task.ID, []string{internal.Configuration.Todoist.NextActionLabel})
						if err != nil {
							internal.Log.Fatalf("Error removing next action label from task %s: %v", task.Content, err)
							break
						}
					}
				}
			}
		}
		internal.Log.Infof("Completed processing of project %s (%s)", project.ID, project.Name)
	}
	internal.Log.Infof("Completed update in %f seconds", time.Since(start).Seconds())
}
