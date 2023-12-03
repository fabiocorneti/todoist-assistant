package internal

import (
	"log"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	LogLevel       string `yaml:"logLevel"`
	UpdateInterval int    `yaml:"updateInterval"`
	Todoist        struct {
		Token                 string `yaml:"token"`
		NextActionLabel       string `yaml:"nextActionLabel"`
		AssignProjectLabel    bool   `yaml:"assignProjectLabel"`
		AssignNextActionLabel bool   `yaml:"assignNextActionLabel"`
		ParentProjectName     string `yaml:"parentProjectName"`
		ProjectsLabelPrefix   string `yaml:"projectsLabelPrefix"`
	} `yaml:"todoist"`
	Jira []JiraConfig `yaml:"jira"`
}

type JiraConfig struct {
	Site               string   `yaml:"site"`
	Username           string   `yaml:"username"`
	Token              string   `yaml:"token"`
	JQL                string   `yaml:"jql"`
	Labels             []string `yaml:"labels"`
	CompletionStatuses []string `yaml:"completionStatuses"`
	Project            string   `yaml:"project"`
	SyncJiraLabels     bool     `yaml:"syncJiraLabels"`
	SyncJiraComponents bool     `yaml:"syncJiraComponents"`
}

var (
	Configuration Config
	Log           = logrus.New()
)

func (cfg *Config) IsTest() bool {
	return cfg.Todoist.Token == "TEST"
}

func (cfg *Config) validate() {
	if cfg.Todoist.Token == "" {
		log.Fatal("Todoist API token is not set")
	}
	if cfg.UpdateInterval <= 0 {
		log.Fatal("Update interval must be greater than 0")
	}
}

func init() {
	configFile := "config.yaml"
	if _, err := os.Stat(configFile); err == nil {
		Configuration, err = readConfig(configFile)
		if err != nil {
			log.Fatalf("Error reading config: %v", err)
		}
	}

	overrideConfigFromEnv(&Configuration)

	setDefaults(&Configuration)

	Configuration.validate()

	if Configuration.LogLevel == "" {
		Configuration.LogLevel = "error"
	}

	if Configuration.LogLevel == "info" {
		Log.SetLevel(logrus.InfoLevel)
	} else if Configuration.LogLevel == "debug" {
		Log.SetLevel(logrus.DebugLevel)
	} else {
		Log.SetLevel(logrus.ErrorLevel)
	}

	if Configuration.Todoist.NextActionLabel == "" {
		Configuration.Todoist.NextActionLabel = "Next Action"
	}
}

func readConfig(path string) (Config, error) {
	var cfg Config
	file, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(file, &cfg)

	return cfg, err
}

func setDefaults(cfg *Config) {
	if cfg.UpdateInterval <= 0 {
		cfg.UpdateInterval = 5
	}
	if cfg.Todoist.ProjectsLabelPrefix == "" {
		cfg.Todoist.ProjectsLabelPrefix = "Projects"
	}
}

func overrideConfigFromEnv(cfg *Config) {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}
	if interval := os.Getenv("UPDATE_INTERVAL"); interval != "" {
		if val, err := strconv.Atoi(interval); err == nil {
			cfg.UpdateInterval = val
		}
	}
	if token := os.Getenv("TODOIST__TOKEN"); token != "" {
		cfg.Todoist.Token = token
	}
	if label := os.Getenv("TODOIST__NEXT_ACTION_LABEL"); label != "" {
		cfg.Todoist.NextActionLabel = label
	}
	if parentProjectName := os.Getenv("TODOIST__PARENT_PROJECT_NAME"); parentProjectName != "" {
		cfg.Todoist.ParentProjectName = parentProjectName
	}
	if prefix := os.Getenv("TODOIST__PROJECTS_LABEL_PREFIX"); prefix != "" {
		cfg.Todoist.ProjectsLabelPrefix = prefix
	}
	if assignNextActionLabel := os.Getenv("TODOIST__ASSIGN_NEXT_ACTION_LABEL"); assignNextActionLabel == "true" {
		cfg.Todoist.AssignNextActionLabel = true
	}
	if assignProjectLabel := os.Getenv("TODOIST__ASSIGN_PROJECT_LABEL"); assignProjectLabel == "true" {
		cfg.Todoist.AssignProjectLabel = true
	}
}
