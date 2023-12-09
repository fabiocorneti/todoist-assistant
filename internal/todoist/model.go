package todoist

type Project struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
}

type Task struct {
	ID        string   `json:"id"`
	Labels    []string `json:"labels"`
	ProjectID string   `json:"project_id,omitempty"`
	Content   string   `json:"content"`
	Order     *int     `json:"order,omitempty"`
	Priority  *int     `json:"priority,omitempty"`
}

type Label struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
