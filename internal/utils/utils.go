package utils

import (
	"fmt"

	"github.com/fabiocorneti/todoist-assistant/internal/config"
	"github.com/fabiocorneti/todoist-assistant/internal/jira"
)

func Contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// HaveSameElements returns true if two slices have the same elements.
// Slices are assumed to not have duplicate elements.
func HaveSameElements(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	elementSet := make(map[string]struct{})

	for _, item := range slice1 {
		elementSet[item] = struct{}{}
	}

	for _, item := range slice2 {
		if _, exists := elementSet[item]; !exists {
			return false
		}
	}

	return true
}

func FormatTodoistTaskContent(jiraConfig config.JiraConfig, issue jira.Issue) string {
	issueURL := fmt.Sprintf("%s/browse/%s", jiraConfig.Site, issue.Key)
	taskContent := fmt.Sprintf("[[%s] %s](%s)", issue.Key, issue.Fields.Summary, issueURL)
	return taskContent
}
