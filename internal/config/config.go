package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	p1 = 4
	p2 = 3
	p3 = 2
	p4 = 1
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
	Jira   []JiraConfig `yaml:"jira"`
	loaded bool
}

type JiraConfig struct {
	Site               string              `yaml:"site"`
	Username           string              `yaml:"username"`
	Token              string              `yaml:"token"`
	JQL                string              `yaml:"jql"`
	Labels             []string            `yaml:"labels"`
	CompletionStatuses []string            `yaml:"completionStatuses"`
	Project            string              `yaml:"project"`
	SyncJiraLabels     bool                `yaml:"syncJiraLabels"`
	SyncJiraComponents bool                `yaml:"syncJiraComponents"`
	PriorityMap        map[string][]string `yaml:"priorityMap"`
}

var (
	configuration Config
	log           = logrus.New()
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

	for _, jiraCfg := range cfg.Jira {
		for key := range jiraCfg.PriorityMap {
			_, err := cfg.ToAPIPriority(key)
			if err != nil {
				log.Fatalf("Invalid key found in PriorityMap: %s. Only p1-p4 are allowed.", key)
			}
		}
	}
}

// ToAPIPriority converts a configuration priority to a Todoist API priority
func (cfg *Config) ToAPIPriority(configPriority string) (int, error) {
	switch configPriority {
	case "p1":
		return p1, nil
	case "p2":
		return p2, nil
	case "p3":
		return p3, nil
	case "p4":
		return p4, nil
	default:
		return 0, fmt.Errorf("unknown priority %s", configPriority)
	}
}

func GetConfiguration() Config {
	if !configuration.loaded {
		loadConfiguration()
	}
	return configuration
}

func GetLogger() *logrus.Logger {
	return log
}

func loadConfiguration() {
	configFile := "config.yaml"
	if _, err := os.Stat(configFile); err == nil {
		configuration, err = readConfig(configFile)
		if err != nil {
			log.Fatalf("Error reading config: %v", err)
		}
	}

	overrideConfigFromEnv(&configuration)

	setDefaults(&configuration)

	configuration.validate()

	if configuration.LogLevel == "" {
		configuration.LogLevel = "error"
	}

	switch configuration.LogLevel {
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	default:
		log.SetLevel(logrus.ErrorLevel)
	}

	if configuration.Todoist.NextActionLabel == "" {
		configuration.Todoist.NextActionLabel = "Next Action"
	}
	configuration.loaded = true
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
