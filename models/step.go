package models

// StepInputModel ...
type StepInputModel struct {
	Key          string   `json:"key"`
	Description  string   `json:"description"`
	DefaultValue string   `json:"default_value"`
	ValueOptions []string `json:"value_options"`
	IsExpand     bool     `json:"is_expand"`
}

// StepVersionModel ...
type StepVersionModel struct {
	ID          string           `json:"step_id"`
	Version     string           `json:"step_version"`
	Description string           `json:"description"`
	Inputs      []StepInputModel `json:"inputs"`
}
