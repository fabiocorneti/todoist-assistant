package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	fields = "key,summary,status,labels,components,priority"
)

type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
		Labels     []string `json:"labels"`
		Components []struct {
			Name string `json:"name"`
		} `json:"components"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
	} `json:"fields"`
}

func FetchJiraIssues(jiraConfig JiraConfig) ([]JiraIssue, error) {
	var allIssues []JiraIssue
	startAt := 0
	maxResults := 50

	for {
		encodedJQL := url.QueryEscape(jiraConfig.JQL)

		requestURL := fmt.Sprintf("%s/rest/api/3/search?jql=%s&startAt=%d&maxResults=%d&fields=%s",
			jiraConfig.Site, encodedJQL, startAt, maxResults, fields)
		req, _ := http.NewRequest("GET", requestURL, nil)
		req.SetBasicAuth(jiraConfig.Username, jiraConfig.Token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error from Jira API: %s", resp.Status)
		}

		body, err := io.ReadAll(io.Reader(resp.Body))
		if err != nil {
			return nil, err
		}
		var response struct {
			Issues     []JiraIssue `json:"issues"`
			Total      int         `json:"total"`
			MaxResults int         `json:"maxResults"`
			StartAt    int         `json:"startAt"`
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		allIssues = append(allIssues, response.Issues...)

		if startAt+len(response.Issues) >= response.Total {
			break
		}
		startAt += len(response.Issues)
	}

	return allIssues, nil
}
