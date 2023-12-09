# Todoist Assistant

This is a simple assistant I use for myself, YMMV.

**Still alpha, might wreak havoc your Todoist installation so be careful**

## What can it do?

- Set a label `Project/<project name>` for each task in a project.
- Set a `Next Action` label in the first actionable task (by order) of a project.
- Fetch issues from Jira sites using JQL queries and create Todoist tasks linked to them.

New Jira tasks will end up in the Inbox and will be completed when the
corresponding Jira issues is closed (as long as the JQL returns it).

## Why?

This is useful to me to:

- Keep track of the lineage of a task when I move it from a topic based project (e.g. `#Projects/#Build a bear` to time based projects (e.g. `#This Week`).
- Set a Next Action in projects (a GTD practice that I don't use much but can be useful when planning).
- Have Jira task references in Todoist (useful for time blocking and for day planning).

I also just like having a simple boilerplate in Go to manipulate Todoist tasks
for all sorts of purposes.

## Usage

Create a yaml file named `config.yaml` with the following contents:

```yaml
todoist:
  token: "YOUR API TOKEN"
updateInterval: 1
```

The program will then process tasks on start and then on every update interval.

If the application cannot find a parent project, it won't process any tasks, but
it can still be used to acquire tasks from Jira.

### Configuration reference

- `logLevel`: The logging level of the application (`debug`, `info`, `error`). Defaults to `error`.
- `updateInterval`: The interval in minutes at which the application processes tasks.
- `todoist`: Todoist configuration.
  - `token`: Your Todoist API token.
  - `assignProjectLabel`: If true, a project label will be assigned to projects. Defaults to `false`.
  - `parentProjectName`: If set, only projects that are child of this project will be assigned project labels.
  - `projectsLabelPrefix`: The prefix used for project labels in Todoist. Defaults to `Projects/`.
  - `assignNextActionLabel`: If true, a Next Action label will be assigned to the first actionable task in projects. Defaults to `false`.
  - `nextActionLabel`: The label used in Todoist to mark the next action. Defaults to `Next Action`.
- `jira`: An array of Jira configurations.

#### Jira configuration

- `site`: The URL of the Jira site (e.g. `https://yourdomain.atlassian.net`).
- `username`: Your Jira username.
- `token`: Your Jira API token.
- `jql`: The JQL query used to fetch issues. If you want tasks to be completed, the JQL should also return closed isseues.
- `labels`: An array of labels that will be added to all Todoist tasks created from Jira issues. Unset by default.
- `completionStatuses`: An array of Jira statuses indicating that an issue has been completed (e.g. `Done`).
- `project`: The name of the project in which new Todoist tasks created from issues will be created. By default they will be created in your Todoist inbox.
- `syncJiraLabels`: A boolean indicating whether to synchronize labels with Jira. Defaults to `false`.
- `syncJiraComponents`: A boolean indicating whether to synchronize components with Jira. Defaults to `false`.
- `priorityMap`: A map between Todoist priorities (p1 to p4) and Jira priority names. Not set by default.

### Full configuration example

```yaml
logLevel: info
update_interval: 1
todoist:
  token: YOUR_TODOIST_TOKEN
  parentProjectName: Projects
  assignProjectLabel: true
  assignNextActionLabel: true
jira:
  - site: "https://yourdomain.atlassian.net"
    username: "spam@smilzo.net"
    token: YOUR_JIRA_TOKEN
    jql: "(status not in ('Backlog', Done, REJECTED) OR (Status in ('Done', REJECTED) AND updated >= -14d))"
    syncJiraComponents: true
    syncJiraLabels: true
    labels:
      - Clients/Mango
      - Jira
    completionStatuses:
      - Done
      - REJECTED
    priorityMap:
      p1:
        - "Highest"
      p2:
        - "High"
      p3:
        - "Medium"
      p4:
        - "Low"
        - "Lowest"
  - site: "https://anotherdomain.atlassian.net"
    username: "spam@smilzo.net"
    token: "YOUR_JIRA_TOKEN"
    jql: "Sprint in openSprints() ORDER BY Rank DESC"
    project: Pear
    labels:
      - Clients/Pear
      - Jira
    completionStatuses:
      - Done
      - Rejected
    priorityMap:
      p1:
        - "Highest"
        - "High"
      p4:
        - "Medium"
        - "Low"
        - "Lowest"
```

#### Environment variable overrides

You can specify the following configuration variables in the environment; if a variable is set both in the environment and in the configuration file, the environment value will take precedence.

- `LOG_LEVEL`: the value for `logLevel`.
- `UPDATE_INTERVAL`: the value for `updateInterval`.
- `TODOIST__TOKEN`: the value for `todoist.token`.
- `TODOIST__PARENT_PROJECT_NAME`: the value for `todoist.parentProjectName`.
-	`TODOIST__ASSIGN_NEXT_ACTION_LABEL`: the value for `todoist.assignNextActionLabel`.
- `TODOIST__ASSIGN_PROJECT_LABEL`: the value for `todoist.assignProjectLabel`.
- `TODOIST__NEXT_ACTION_LABEL`: the value for `todoist.nextActionLabel`
- `TODOIST__PROJECTS_LABEL_PREFIX`: the value for `todoist.projectsLabelPrefix`

If you need Jira integration you'll need to mount a proper configuration file as the configuration would be too complex for environment variables.

In general the configuration file is the recommended way to go unless you are really forced to use only environment variables.

## How to run with Docker

A Docker image built from the main branch is available at https://hub.docker.com/repository/docker/corneti/todoist-assistant ; no
releases at the moment other than `latest`, so if you want the latest version
you'll need to pull from the hub.

To use the image you can simply mount the configuration file at `/config.yaml`.

## Notes

- The program is currently designed to be stateless, so it will reprocess all tasks on every run.
- Labels are assigned to tasks by name, which means they will end up in your Shared labels.

## Known limitations

- Alpha quality, still being tested, mostly in a works for me fashion.
- No E2E tests, scarce unit tests.
- Does not work yet on recursive project structures when a parent project name is specified.
- Will abort on any request error.
